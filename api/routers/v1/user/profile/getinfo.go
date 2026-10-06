package profile

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
)

func GetUserInfoHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	targetID, ok := parseUserID(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	user, err := servants.UserServant.QueryByID(targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.UserNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}

	// 目标用户处于账号封禁状态（不允许登录）时，返回 40304
	if user.Status == consts.UserStatusBanned {
		return resp.TargetPunished(ctx)
	}

	blockStatus, err := servants.BlockServant.GetBlockStatus(ctx.Context(), uid, targetID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	// 目标用户当前生效中的权限封禁位掩码，供前端展示具体受限功能
	punishmentMask, err := servants.UserServant.QueryActivePunishmentMask(targetID, time.Now())
	if err != nil {
		return resp.ServerError(ctx)
	}

	// 手机号是账号凭据信息，只在「查询自己」时返回，且一律返回打码值；
	// 查询他人资料时字段为空并由 omitempty 省略，避免任何形式的手机号泄露。
	phone := ""
	if targetID == uid {
		phone = MaskedSessionPhone(ctx)
	}

	return resp.UserInfoOK(ctx, resp.UserInfo{
		User:           *user,
		BlockStatus:    blockStatus,
		PunishmentMask: punishmentMask,
		Phone:          phone,
	})
}

func parseUserID(ctx fiber.Ctx) (uint32, bool) {
	value := strings.TrimSpace(ctx.Params("uid"))
	if value == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(value, 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint32(id), true
}
