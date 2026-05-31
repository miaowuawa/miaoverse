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

func LoginBySMSSingleAccount(c fiber.Ctx, phone string, region int, id uint64) error {
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

func LoginBySMSMultipleChoices(c fiber.Ctx, phone string, region int) error {
	if err := loginUniversal(c); err != nil {
		return err
	}

	sess := session.FromContext(c)
	sess.Delete(SessionUID)
	sess.Set(SessionPendingLoginPhone, phone)
	sess.Set(SessionPendingLoginRegion, region)
	return nil
}

func PendingSMSLogin(c fiber.Ctx) (string, int, bool) {
	sess := session.FromContext(c)
	if sess == nil {
		return "", 0, false
	}

	phone, ok := sess.Get(SessionPendingLoginPhone).(string)
	if !ok || phone == "" {
		return "", 0, false
	}

	region, ok := sessionValueToInt(sess.Get(SessionPendingLoginRegion))
	if !ok {
		return "", 0, false
	}

	return phone, region, true
}

func CurrentUID(c fiber.Ctx) (uint64, bool) {
	sess := session.FromContext(c)
	if sess == nil {
		return 0, false
	}
	return sessionValueToUint64(sess.Get(SessionUID))
}

func sessionValueToInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	default:
		return 0, false
	}
}

func sessionValueToUint64(value any) (uint64, bool) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int8:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int16:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int32:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	default:
		return 0, false
	}
}
