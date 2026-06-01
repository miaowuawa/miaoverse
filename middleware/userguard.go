package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/UserSession"
	"miaoverse/service/i18n"
)

const (
	userLocalKey = "miaoverse.user"
	uidLocalKey  = "miaoverse.uid"
)

const (
	activeUserStatus   uint8 = 1
	bannedUserStatus   uint8 = 2
	closedUserStatus   uint8 = 3
	disabledUserStatus uint8 = 4
)

type UserContext struct {
	UID         uint64
	User        *modeluser.User
	servants    *server.Servants
	credentials map[uint8]bool
}

type UserCheck func(fiber.Ctx, *UserContext) *userGuardReject

type userGuardReject struct {
	status int
	msg    i18n.MessageKey
}

func RequireUser(servants *server.Servants, checks ...UserCheck) fiber.Handler {
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

		userCtx := &UserContext{
			UID:         uid,
			User:        u,
			servants:    servants,
			credentials: map[uint8]bool{},
		}
		ctx.Locals(uidLocalKey, uid)
		ctx.Locals(userLocalKey, userCtx)

		for _, check := range checks {
			if check == nil {
				continue
			}
			if reject := check(ctx, userCtx); reject != nil {
				return rejectUserGuard(ctx, reject.status, reject.msg)
			}
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
	userCtx, ok := ctx.Locals(userLocalKey).(*UserContext)
	if !ok || userCtx == nil || userCtx.User == nil {
		return nil, false
	}
	return userCtx.User, true
}

func AccountActive() UserCheck {
	return func(_ fiber.Ctx, userCtx *UserContext) *userGuardReject {
		switch userCtx.User.Status {
		case activeUserStatus:
			return nil
		case bannedUserStatus:
			return &userGuardReject{status: fiber.StatusForbidden, msg: i18n.ErrAccountBanned}
		case closedUserStatus:
			return &userGuardReject{status: fiber.StatusForbidden, msg: i18n.ErrAccountClosed}
		case disabledUserStatus:
			return &userGuardReject{status: fiber.StatusForbidden, msg: i18n.ErrAccountDisabled}
		default:
			return &userGuardReject{status: fiber.StatusForbidden, msg: i18n.ErrAccountUnavailable}
		}
	}
}

func PhoneBound() UserCheck {
	return CredentialBound(consts.Phone, i18n.ErrPhoneNotBound)
}

func PasswordSet() UserCheck {
	return CredentialBound(consts.Password, i18n.ErrPasswordNotSet)
}

func Certified() UserCheck {
	return CredentialBound(consts.ThirdPartyWebAuthn, i18n.ErrCertificationRequired)
}

func CredentialBound(credType uint8, msg i18n.MessageKey) UserCheck {
	return func(_ fiber.Ctx, userCtx *UserContext) *userGuardReject {
		ok, err := userCtx.hasCredential(credType)
		if err != nil {
			return &userGuardReject{status: fiber.StatusInternalServerError, msg: i18n.ErrServerContactAdmin}
		}
		if !ok {
			return &userGuardReject{status: fiber.StatusForbidden, msg: msg}
		}
		return nil
	}
}

func (u *UserContext) hasCredential(credType uint8) (bool, error) {
	if ok, cached := u.credentials[credType]; cached {
		return ok, nil
	}

	ok, err := u.servants.UserServant.HasCredential(u.UID, credType)
	if err != nil {
		return false, err
	}
	u.credentials[credType] = ok
	return ok, nil
}

func rejectUserGuard(ctx fiber.Ctx, status int, msg i18n.MessageKey) error {
	return ctx.Status(status).JSON(resp.CodeWithMsg{
		Code: status,
		Msg:  i18n.Message(ctx, msg),
	})
}
