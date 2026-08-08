package middleware

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	modelinteracts "miaoverse/model/dao/interacts"
	modelmoment "miaoverse/model/dao/moment"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/UserBlock"
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
	AllowSelf        bool // 允许目标为自己（如回复自己的评论、评论自己的动态）
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
		if targetID == 0 || (!config.AllowSelf && targetID == uid) {
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

		ctx.Locals(consts.BlockTargetLocalKey, targetID)
		return ctx.Next()
	}
}

// BlockTarget 读取 RequireNoBlock 解析出的目标用户 ID。
func BlockTarget(ctx fiber.Ctx) (uint32, bool) {
	if v, ok := ctx.Locals(consts.BlockTargetLocalKey).(uint32); ok && v != 0 {
		return v, true
	}
	return 0, false
}

// BlockMoment 读取 RequireNoBlock 的 Resolver 存入的动态对象（点赞/评论场景复用，避免重复查询）。
func BlockMoment(ctx fiber.Ctx) (*modelmoment.Moment, bool) {
	if m, ok := ctx.Locals(consts.BlockMomentLocalKey).(*modelmoment.Moment); ok && m != nil {
		return m, true
	}
	return nil, false
}

// BlockComment 读取 RequireNoBlock 的 Resolver 存入的被回复评论对象（回复评论/楼中楼对话场景）。
func BlockComment(ctx fiber.Ctx) (*modelinteracts.Comment, bool) {
	if c, ok := ctx.Locals(consts.BlockCommentLocalKey).(*modelinteracts.Comment); ok && c != nil {
		return c, true
	}
	return nil, false
}

// BlockCommentRoot 读取 RequireNoBlock 的 Resolver 存入的楼中楼首条评论对象（用于互动记录 reference_id）。
func BlockCommentRoot(ctx fiber.Ctx) (*modelinteracts.Comment, bool) {
	if c, ok := ctx.Locals(consts.BlockCommentRootKey).(*modelinteracts.Comment); ok && c != nil {
		return c, true
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
	if moment.Status != consts.MomentStatusNormal {
		return 0, errBlockTargetMissing
	}

	ctx.Locals(consts.BlockMomentLocalKey, moment)
	return moment.UserID, nil
}

// ResolveCommentAuthor 回复评论/楼中楼对话场景的目标解析器：按路径参数 :id 查评论，返回评论作者 ID 并缓存评论对象。
func ResolveCommentAuthor(ctx fiber.Ctx, servants *server.Servants) (uint32, error) {
	comment, err := loadCommentByPathID(ctx, servants)
	if err != nil {
		return 0, err
	}
	return comment.UserID, nil
}

// ResolveCommentMomentAuthor 回复评论/楼中楼对话场景的目标解析器：按路径参数 :id 查评论，
// 沿 target 链上溯到所属动态，返回动态作者 ID，并缓存被回复评论、楼中楼首条评论与动态对象。
func ResolveCommentMomentAuthor(ctx fiber.Ctx, servants *server.Servants) (uint32, error) {
	comment, err := loadCommentByPathID(ctx, servants)
	if err != nil {
		return 0, err
	}

	root, moment, err := resolveCommentRoot(comment, servants)
	if err != nil {
		return 0, err
	}
	ctx.Locals(consts.BlockCommentRootKey, root)
	ctx.Locals(consts.BlockMomentLocalKey, moment)
	return moment.UserID, nil
}

// loadCommentByPathID 从路径参数 :id 解析评论，校验存在且状态正常，并缓存评论对象。
// 已缓存时直接复用（RequireNoBlock 可叠加多个 Resolver 时避免重复查询）。
func loadCommentByPathID(ctx fiber.Ctx, servants *server.Servants) (*modelinteracts.Comment, error) {
	if comment, ok := BlockComment(ctx); ok {
		return comment, nil
	}

	id, err := strconv.ParseUint(strings.TrimSpace(ctx.Params("id")), 10, 64)
	if err != nil || id == 0 {
		return nil, errBlockTargetMissing
	}
	comment, err := servants.InteractsServant.QueryCommentByID(id)
	if err != nil || comment.Status != consts.CommentStatusNormal {
		return nil, errBlockTargetMissing
	}

	ctx.Locals(consts.BlockCommentLocalKey, comment)
	return comment, nil
}

// resolveCommentRoot 沿评论 target 链上溯：找到楼中楼首条评论与所属动态。
// 评论本身 target_type=moment 时，该评论即首条评论。
func resolveCommentRoot(comment *modelinteracts.Comment, servants *server.Servants) (*modelinteracts.Comment, *modelmoment.Moment, error) {
	root := comment
	for depth := 0; depth < consts.CommentChainMaxDepth; depth++ {
		if root.TargetType == consts.CommentTargetMoment {
			moment, err := servants.ContentServant.QueryMomentByID(root.TargetID)
			if err != nil || moment.Status != consts.MomentStatusNormal {
				return nil, nil, errBlockTargetMissing
			}
			return root, moment, nil
		}
		if root.TargetType != consts.CommentTargetComment {
			return nil, nil, errBlockTargetMissing
		}
		parent, err := servants.InteractsServant.QueryCommentByID(root.TargetID)
		if err != nil || parent.Status != consts.CommentStatusNormal {
			return nil, nil, errBlockTargetMissing
		}
		root = parent
	}
	return nil, nil, errBlockTargetMissing
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
