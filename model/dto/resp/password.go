package resp

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/service/i18n"
)

// CodeWithMsgPassword 密码相关响应：has_password 表示当前账号是否已设置密码，
// 供前端展示「设置密码 / 修改密码」。不返回任何密码哈希或盐值。
type CodeWithMsgPassword struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg"`
	HasPassword bool   `json:"has_password"`
}

// PasswordStatus 密码状态查询成功响应。
func PasswordStatus(ctx fiber.Ctx, hasPassword bool) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgPassword{
		Code:        fiber.StatusOK,
		Msg:         i18n.Message(ctx, i18n.OKPasswordStatus),
		HasPassword: hasPassword,
	})
}

// PasswordUpdated 密码设置/修改成功响应。
func PasswordUpdated(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgPassword{
		Code:        fiber.StatusOK,
		Msg:         i18n.Message(ctx, i18n.OKPasswordUpdated),
		HasPassword: true,
	})
}
