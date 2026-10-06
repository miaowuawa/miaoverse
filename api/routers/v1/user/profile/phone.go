package profile

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/service/UserProfile"
	"miaoverse/service/UserSession"
)

// MaskedSessionPhone 返回当前登录会话绑定手机号的脱敏展示值（如 "+86 138****8000"）。
//
// 手机号保存在 user_credentials 表且当前不支持更改，因此这里只从会话读取、只用于展示，
// 全项目统一走本函数，保证「本人信息」相关接口（/user/me、/user/info、本人资料查询）
// 返回的脱敏格式一致，也不会出现某个接口漏打码、返回完整号码的情况。
// 会话没有绑定手机号（例如非短信登录）时返回空字符串，调用方配合 omitempty 省略字段。
func MaskedSessionPhone(ctx fiber.Ctx) string {
	phone, region, ok := UserSession.CurrentPhoneRegion(ctx)
	if !ok {
		return ""
	}
	return UserProfile.FormatMaskedPhone(phone, region)
}
