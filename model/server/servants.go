package server

import (
	"github.com/go-playground/validator/v10"
	fiberstoreredis "github.com/gofiber/storage/redis/v3"
	"miaoverse/dao/user"
	"miaoverse/service/security/sms/codemanager"
	"miaoverse/service/security/sms/smsbao"
)

type Servants struct {
	FiberSessionStorage *fiberstoreredis.Storage
	SmsServant          *smsbao.SmsBaoServant
	CodeManager         *codemanager.CodeManager
	Validator           *validator.Validate
	UserServant         *user.UserDAO
}
