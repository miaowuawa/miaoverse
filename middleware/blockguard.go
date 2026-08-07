package middleware

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	modelmoment "miaoverse/model/dao/moment"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/UserBlock"
)

const (
	blockTargetLocalKey = "miaoverse.block.target"
	blockMomentLocalKey = "miaoverse.block.moment"
)

var errBlockTargetMissing = errors.New("block target is missing")

// BlockTargetResolver 自定义目标用户 ID 解析器（如点赞/评论需先查动态再取作者）。
// 返回的 targetID 会写入 ctx.Locals，handler 可通过 BlockTarget(ctx) 复用。
type BlockTargetResolver func(ctx fiber.Ctx, servants *server.Servants) (uint32, error)

// BlockGuardConfig 拉黑校验中间件配置。
// PathParam 与 BodyField 二选一；都不填时必须提供 Resolver。
type BlockGuardConfig struct {
	PathParam        string
	BodyField        string
	Resolver         BlockTargetResolver
	CheckMuteUnwatch bool // 额外校验目标被自己屏蔽/不想看（关注场景）
}

// RequireNoBlock 拉黑校验中间件：查看者与目标用户存在任意一方拉黑关系时拒绝（40301）。
// 目标用户 ID 来源：PathParam（如 /users/:uid/...）、BodyField（如 target）、或 Resolver。
func RequireNoBlock(servants *server.Servants, config BlockGuardConfig) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		uid, ok := CurrentUID(ctx)
		if !ok {
			return resp.Unauthorized(ctx)
		}

		targetID, err := resolveBlockTarget(ctx, servants, config)
		if err != nil {
			return resp.BadRequest(ctx)
		}
		if targetID == 0 || targetID == uid {
			return resp.BadRequest(ctx)
		}

		blocked, err := servants.BlockServant.IsBlockedEither(ctx.Context(), uid, targetID)
		if err != nil {
			return resp.ServerError(ctx)
		}
		if blocked {
			return resp.Blocked(ctx)
		}

		if config.CheckMuteUnwatch {
			for _, blockType := range []UserBlock.BlockType{UserBlock.BlockTypeMute, UserBlock.BlockTypeUnwatch} {
				contained, err := servants.BlockServant.Contains(ctx.Context(), uid, blockType, targetID)
				if err != nil {
					return resp.ServerError(ctx)
				}
				if contained {
					return resp.Blocked(ctx)
				}
			}
		}

		ctx.Locals(blockTargetLocalKey, targetID)
		return ctx.Next()
	}
}

// BlockTarget 读取 RequireNoBlock 解析出的目标用户 ID。
func BlockTarget(ctx fiber.Ctx) (uint32, bool) {
	if v, ok := ctx.Locals(blockTargetLocalKey).(uint32); ok && v != 0 {
		return v, true
	}
	return 0, false
}

// BlockMoment 读取 RequireNoBlock 的 Resolver 存入的动态对象（点赞/评论场景复用，避免重复查询）。
func BlockMoment(ctx fiber.Ctx) (*modelmoment.Moment, bool) {
	if m, ok := ctx.Locals(blockMomentLocalKey).(*modelmoment.Moment); ok && m != nil {
		return m, true
	}
	return nil, false
}

// ResolveMomentAuthor 点赞/评论场景的目标解析器：按 body 中 moment_id 查动态，返回作者 ID 并缓存动态对象。
func ResolveMomentAuthor(ctx fiber.Ctx, servants *server.Servants) (uint32, error) {
	var body struct {
		MomentID uint64 `json:"moment_id"`
	}
	if err := json.Unmarshal(ctx.Body(), &body); err != nil || body.MomentID == 0 {
		return 0, errBlockTargetMissing
	}

	moment, err := servants.ContentServant.QueryMomentByID(body.MomentID)
	if err != nil {
		return 0, errBlockTargetMissing
	}
	if moment.Status != modelmoment.MomentStatusNormal {
		return 0, errBlockTargetMissing
	}

	ctx.Locals(blockMomentLocalKey, moment)
	return moment.UserID, nil
}

func resolveBlockTarget(ctx fiber.Ctx, servants *server.Servants, config BlockGuardConfig) (uint32, error) {
	if config.Resolver != nil {
		return config.Resolver(ctx, servants)
	}
	if config.PathParam != "" {
		value := strings.TrimSpace(ctx.Params(config.PathParam))
		if value == "" {
			return 0, errBlockTargetMissing
		}
		id, err := strconv.ParseUint(value, 10, 32)
		if err != nil || id == 0 {
			return 0, errBlockTargetMissing
		}
		return uint32(id), nil
	}
	if config.BodyField != "" {
		var body map[string]json.RawMessage
		if err := json.Unmarshal(ctx.Body(), &body); err != nil {
			return 0, errBlockTargetMissing
		}
		raw, ok := body[config.BodyField]
		if !ok {
			return 0, errBlockTargetMissing
		}
		var id uint64
		if err := json.Unmarshal(raw, &id); err != nil || id == 0 {
			return 0, errBlockTargetMissing
		}
		return uint32(id), nil
	}
	return 0, errBlockTargetMissing
}
