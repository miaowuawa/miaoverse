package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/UserCheck"
	"miaoverse/service/UserSession"
	"miaoverse/service/i18n"
)

const (
	userLocalKey = "miaoverse.user"
	uidLocalKey  = "miaoverse.uid"
)

func RequireUser(servants *server.Servants, checks ...UserCheck.Check) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		uid, ok := UserSession.CurrentUID(ctx)
		if !ok {
			return rejectUserGuard(ctx, fiber.StatusUnauthorized, i18n.ErrUnauthorized)
		}

		u, err := servants.UserServant.QueryByID(uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return rejectUserGuard(ctx, fiber.StatusUnauthorized, i18n.ErrUnauthorized)
			}
			return rejectUserGuard(ctx, fiber.StatusInternalServerError, i18n.ErrServerContactAdmin)
		}

		userCtx := UserCheck.NewContext(uid, u, servants)
		ctx.Locals(uidLocalKey, uid)
		ctx.Locals(userLocalKey, userCtx)

		for _, check := range checks {
			if check == nil {
				continue
			}
			result := check(userCtx)
			if result.Passed() {
				continue
			}
			if result.Err != nil {
				return rejectUserGuard(ctx, fiber.StatusInternalServerError, i18n.ErrServerContactAdmin)
			}
			return rejectUserGuard(ctx, fiber.StatusForbidden, failureMessage(result.Failure))
		}

		return ctx.Next()
	}
}

func CurrentUID(ctx fiber.Ctx) (uint64, bool) {
	if uid, ok := ctx.Locals(uidLocalKey).(uint64); ok {
		return uid, true
	}
	return UserSession.CurrentUID(ctx)
}

func CurrentUser(ctx fiber.Ctx) (*modeluser.User, bool) {
	userCtx, ok := ctx.Locals(userLocalKey).(*UserCheck.Context)
	if !ok || userCtx == nil || userCtx.User == nil {
		return nil, false
	}
	return userCtx.User, true
}

func failureMessage(failure UserCheck.Failure) i18n.MessageKey {
	switch failure {
	case UserCheck.AccountBanned:
		return i18n.ErrAccountBanned
	case UserCheck.AccountClosed:
		return i18n.ErrAccountClosed
	case UserCheck.AccountDisabled:
		return i18n.ErrAccountDisabled
	case UserCheck.PhoneNotBound:
		return i18n.ErrPhoneNotBound
	case UserCheck.PasswordNotSet:
		return i18n.ErrPasswordNotSet
	case UserCheck.CertificationRequired:
		return i18n.ErrCertificationRequired
	default:
		return i18n.ErrAccountUnavailable
	}
}

func rejectUserGuard(ctx fiber.Ctx, status int, msg i18n.MessageKey) error {
	return resp.JSON(ctx, status, msg)
}
