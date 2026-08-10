package profile

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/i18n"
)

// MeHandler 返回当前登录用户的基础信息，供前端恢复登录态。
func MeHandler(ctx fiber.Ctx, servants *server.Servants) error {
	user, ok := middleware.CurrentUser(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	return ctx.Status(fiber.StatusOK).JSON(resp.CodeWithMsgUser{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKUserInfoFetched),
		User: *user,
	})
}
