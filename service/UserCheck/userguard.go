package UserCheck

import (
	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/server"
)

type Failure string

const (
	Pass                  Failure = ""
	AccountBanned         Failure = "account_banned"
	AccountClosed         Failure = "account_closed"
	AccountDisabled       Failure = "account_disabled"
	AccountUnavailable    Failure = "account_unavailable"
	PhoneNotBound         Failure = "phone_not_bound"
	PasswordNotSet        Failure = "password_not_set"
	CertificationRequired Failure = "certification_required"
)

type Context struct {
	UID         uint32
	User        *modeluser.User
	servants    *server.Servants
	credentials map[uint8]bool
}

type Result struct {
	Failure Failure
	Err     error
}

type Check func(*Context) Result

func NewContext(uid uint32, user *modeluser.User, servants *server.Servants) *Context {
	return &Context{
		UID:         uid,
		User:        user,
		servants:    servants,
		credentials: map[uint8]bool{},
	}
}

func OK() Result {
	return Result{Failure: Pass}
}

func Failed(failure Failure) Result {
	return Result{Failure: failure}
}

func (r Result) Passed() bool {
	return r.Err == nil && r.Failure == Pass
}

func AccountActive() Check {
	return func(ctx *Context) Result {
		switch ctx.User.Status {
		case consts.UserStatusActive:
			return OK()
		case consts.UserStatusBanned:
			return Failed(AccountBanned)
		case consts.UserStatusClosed:
			return Failed(AccountClosed)
		case consts.UserStatusDisabled:
			return Failed(AccountDisabled)
		default:
			return Failed(AccountUnavailable)
		}
	}
}

func PhoneBound() Check {
	return CredentialBound(consts.Phone, PhoneNotBound)
}

func PasswordSet() Check {
	return CredentialBound(consts.Password, PasswordNotSet)
}

func Certified() Check {
	return CredentialBound(consts.ThirdPartyWebAuthn, CertificationRequired)
}

func CredentialBound(credType uint8, failure Failure) Check {
	return func(ctx *Context) Result {
		ok, err := ctx.hasCredential(credType)
		if err != nil {
			return Result{Err: err}
		}
		if !ok {
			return Failed(failure)
		}
		return OK()
	}
}

func (ctx *Context) hasCredential(credType uint8) (bool, error) {
	if ok, cached := ctx.credentials[credType]; cached {
		return ok, nil
	}

	ok, err := ctx.servants.UserServant.HasCredential(ctx.UID, credType)
	if err != nil {
		return false, err
	}
	ctx.credentials[credType] = ok
	return ok, nil
}
