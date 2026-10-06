package profile

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/user/avatarreq"
	"miaoverse/model/server"
	"miaoverse/service/UserProfile"
)

// UpdateAvatarHandler 设置当前登录用户的头像。
// 头像文件必须是当前用户自己的 active 图片文件，且必须公开（permission=0），
// 否则头像无法被所有人查看（头像展示不受拉黑/屏蔽影响）。
// 修改头像需要未被封禁 PermAvatar 权限位，否则返回 40302。
//
// 文件归属/类型/公开性校验与 PATCH /api/v1/user/info 的 avatar 字段共用
// service/UserProfile，两条路径校验强度一致，不存在绕过归属校验的可能。
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

	field, err := UserProfile.AvatarField(servants.UserServant, uid, req.AvatarUUID)
	if err != nil {
		return avatarError(ctx, err)
	}

	if _, err := UserProfile.Update(servants.UserServant, uid, []UserProfile.Field{field}, time.Now()); err != nil {
		if errors.Is(err, UserProfile.ErrPunished) {
			return resp.Punished(ctx)
		}
		return resp.ServerError(ctx)
	}

	avatarUUID, _ := field.Value.(string)
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
