package resp

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/service/i18n"
)

// MaskClosedAccount 将已注销账号的展示字段统一替换为固定文案，
// 避免在资料、关系列表等场景泄露注销前的用户名与个性签名。
// 停用（UserStatusDisabled）等其他状态不处理。
func MaskClosedAccount(ctx fiber.Ctx, u *modeluser.User) {
	MaskClosedAccountByLang(i18n.LanguageFromCtx(ctx), u)
}

// MaskClosedAccountByLang 与 MaskClosedAccount 语义一致，按指定语言取打码文案。
// 供无请求上下文的异步出站场景（如 SSE 通知推送）复用同一套打码逻辑。
func MaskClosedAccountByLang(lang string, u *modeluser.User) {
	if u == nil || u.Status != consts.UserStatusClosed {
		return
	}
	u.Username = i18n.MessageByLang(lang, i18n.UserClosedUsername)
	u.Bio = i18n.MessageByLang(lang, i18n.UserClosedBio)
}
