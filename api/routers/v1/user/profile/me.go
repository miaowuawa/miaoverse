package profile

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/i18n"
)

// MeHandler 返回当前登录用户的基础信息，供前端恢复登录态与渲染个人资料页。
// 额外返回打码后的绑定手机号：手机号取自登录会话（无需查库），且当前不支持更改，
// 页面只做只读展示，避免用户误以为可以换绑。
func MeHandler(ctx fiber.Ctx, servants *server.Servants) error {
	user, ok := middleware.CurrentUser(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	resp.MaskClosedAccount(ctx, user)

	return ctx.Status(fiber.StatusOK).JSON(resp.CodeWithMsgUser{
		Code:  fiber.StatusOK,
		Msg:   i18n.Message(ctx, i18n.OKUserInfoFetched),
		User:  *user,
		Phone: MaskedSessionPhone(ctx),
	})
}
