package UserSession

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"miaoverse/consts"
)

func LoginBySMSSingleAccount(c fiber.Ctx, phone string, region uint16, id uint32) error {
	if err := loginUniversal(c); err != nil {
		return err
	}

	sess := session.FromContext(c)
	sess.Delete(consts.SessionPendingLoginPhone)
	sess.Delete(consts.SessionPendingLoginRegion)
	sess.Set(consts.SessionPhone, phone)
	sess.Set(consts.SessionRegion, region)
	sess.Set(consts.SessionUID, id)
	return nil
}

func LoginBySMSMultipleChoices(c fiber.Ctx, phone string, region uint16) error {
	if err := loginUniversal(c); err != nil {
		return err
	}

	sess := session.FromContext(c)
	sess.Delete(consts.SessionUID)
	sess.Set(consts.SessionPendingLoginPhone, phone)
	sess.Set(consts.SessionPendingLoginRegion, region)
	return nil
}

func PendingSMSLogin(c fiber.Ctx) (string, uint16, bool) {
	sess := session.FromContext(c)
	if sess == nil {
		return "", 0, false
	}

	phone, ok := sess.Get(consts.SessionPendingLoginPhone).(string)
	if !ok || phone == "" {
		return "", 0, false
	}

	region, ok := sessionValueToUint16(sess.Get(consts.SessionPendingLoginRegion))
	if !ok {
		return "", 0, false
	}

	return phone, region, true
}

func CurrentUID(c fiber.Ctx) (uint32, bool) {
	sess := session.FromContext(c)
	if sess == nil {
		return 0, false
	}
	return sessionValueToUint32(sess.Get(consts.SessionUID))
}

// CurrentPhoneRegion 返回当前会话绑定的手机号与区号（仅短信登录会话写入）。
func CurrentPhoneRegion(c fiber.Ctx) (string, uint16, bool) {
	sess := session.FromContext(c)
	if sess == nil {
		return "", 0, false
	}

	phone, ok := sess.Get(consts.SessionPhone).(string)
	if !ok || phone == "" {
		return "", 0, false
	}

	region, ok := sessionValueToUint16(sess.Get(consts.SessionRegion))
	if !ok {
		return "", 0, false
	}

	return phone, region, true
}

// SwitchAccount 将当前会话切换到同一手机号下绑定的另一个账号，并重新生成 session ID 防止会话固定。
func SwitchAccount(c fiber.Ctx, phone string, region uint16, id uint32) error {
	if err := loginUniversal(c); err != nil {
		return err
	}

	sess := session.FromContext(c)
	sess.Set(consts.SessionPhone, phone)
	sess.Set(consts.SessionRegion, region)
	sess.Set(consts.SessionUID, id)
	return nil
}

// Logout 销毁当前会话并清除登录态 cookie。
// 未登录时也返回 nil，保证退出登录接口幂等。
func Logout(c fiber.Ctx) error {
	sess := session.FromContext(c)
	if sess == nil {
		return nil
	}
	return sess.Destroy()
}

func sessionValueToUint16(value any) (uint16, bool) {
	switch v := value.(type) {
	case int:
		if v < 0 || v > 65535 {
			return 0, false
		}
		return uint16(v), true
	case int8:
		if v < 0 {
			return 0, false
		}
		return uint16(v), true
	case int16:
		if v < 0 {
			return 0, false
		}
		return uint16(v), true
	case int32:
		if v < 0 || v > 65535 {
			return 0, false
		}
		return uint16(v), true
	case int64:
		if v < 0 || v > 65535 {
			return 0, false
		}
		return uint16(v), true
	case uint:
		if v > 65535 {
			return 0, false
		}
		return uint16(v), true
	case uint8:
		return uint16(v), true
	case uint16:
		return v, true
	case uint32:
		if v > 65535 {
			return 0, false
		}
		return uint16(v), true
	case uint64:
		if v > 65535 {
			return 0, false
		}
		return uint16(v), true
	default:
		return 0, false
	}
}

func sessionValueToUint32(value any) (uint32, bool) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, false
		}
		return uint32(v), true
	case int8:
		if v < 0 {
			return 0, false
		}
		return uint32(v), true
	case int16:
		if v < 0 {
			return 0, false
		}
		return uint32(v), true
	case int32:
		if v < 0 {
			return 0, false
		}
		return uint32(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint32(v), true
	case uint:
		return uint32(v), true
	case uint8:
		return uint32(v), true
	case uint16:
		return uint32(v), true
	case uint32:
		return v, true
	case uint64:
		return uint32(v), true
	default:
		return 0, false
	}
}
