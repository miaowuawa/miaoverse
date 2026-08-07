package UserSession

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

const (
	SessionPhone              = "Phone"
	SessionRegion             = "Region"
	SessionUID                = "UID"
	SessionPendingLoginPhone  = "PendingLoginPhone"
	SessionPendingLoginRegion = "PendingLoginRegion"
)

func LoginBySMSSingleAccount(c fiber.Ctx, phone string, region uint16, id uint32) error {
	if err := loginUniversal(c); err != nil {
		return err
	}

	sess := session.FromContext(c)
	sess.Delete(SessionPendingLoginPhone)
	sess.Delete(SessionPendingLoginRegion)
	sess.Set(SessionPhone, phone)
	sess.Set(SessionRegion, region)
	sess.Set(SessionUID, id)
	return nil
}

func LoginBySMSMultipleChoices(c fiber.Ctx, phone string, region uint16) error {
	if err := loginUniversal(c); err != nil {
		return err
	}

	sess := session.FromContext(c)
	sess.Delete(SessionUID)
	sess.Set(SessionPendingLoginPhone, phone)
	sess.Set(SessionPendingLoginRegion, region)
	return nil
}

func PendingSMSLogin(c fiber.Ctx) (string, uint16, bool) {
	sess := session.FromContext(c)
	if sess == nil {
		return "", 0, false
	}

	phone, ok := sess.Get(SessionPendingLoginPhone).(string)
	if !ok || phone == "" {
		return "", 0, false
	}

	region, ok := sessionValueToUint16(sess.Get(SessionPendingLoginRegion))
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
	return sessionValueToUint32(sess.Get(SessionUID))
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
