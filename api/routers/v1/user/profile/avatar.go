package profile

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/user/avatarreq"
	"miaoverse/model/server"
	"miaoverse/service/UserFile"
)

// UpdateAvatarHandler 设置当前登录用户的头像。
// 头像文件必须是当前用户自己的 active 图片文件，且必须公开（permission=0），
// 否则头像无法被所有人查看（头像展示不受拉黑/屏蔽影响）。
// 修改头像需要未被封禁 PermAvatar 权限位，否则返回 40302。
func UpdateAvatarHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	req := &avatarreq.UpdateAvatar{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	avatarUUID := strings.TrimSpace(req.AvatarUUID)
	if avatarUUID == "" {
		return resp.BadRequest(ctx)
	}

	punished, err := servants.UserServant.HasActivePunishment(uid, consts.PermAvatar, time.Now())
	if err != nil {
		return resp.ServerError(ctx)
	}
	if punished {
		return resp.Punished(ctx)
	}

	file, ok := UserFile.ValidateAvatarUUID(servants.UserServant, uid, avatarUUID)
	if !ok {
		return resp.BadRequest(ctx)
	}
	// 头像必须公开分享：非公开文件无法被所有人查看，且头像展示不受拉黑/屏蔽影响
	if file.Permission != consts.FilePermissionPublic {
		return resp.BadRequest(ctx)
	}

	if _, err := servants.UserServant.UpdateProfile(uid, map[string]any{"avatar": avatarUUID}); err != nil {
		return resp.ServerError(ctx)
	}

	return resp.AvatarUpdated(ctx, avatarUUID)
}

// GetAvatarHandler 获取任意用户当前头像的文件 UUID。
// 头像为公开可见文件，展示不受拉黑/屏蔽/账号封禁影响，因此不做任何关系校验。
func GetAvatarHandler(ctx fiber.Ctx, servants *server.Servants) error {
	targetID, ok := parseUserID(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	user, err := servants.UserServant.QueryByID(targetID)
	if err != nil {
		return resp.UserNotFound(ctx)
	}

	return resp.AvatarFetched(ctx, user.Avatar)
}
