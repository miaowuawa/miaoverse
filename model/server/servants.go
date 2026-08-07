package server

import (
	"github.com/go-playground/validator/v10"
	fiberstoreredis "github.com/gofiber/storage/redis/v3"
	"miaoverse/dao/content"
	"miaoverse/dao/user"
	storages3 "miaoverse/service/s3"
	"miaoverse/service/security/sms/codemanager"
	"miaoverse/service/security/sms/smsbao"
)

type Servants struct {
	FiberSessionStorage *fiberstoreredis.Storage
	SmsServant          *smsbao.SmsBaoServant
	CodeManager         *codemanager.CodeManager
	Validator           *validator.Validate
	UserServant         *user.UserDAO
	ContentServant      *content.ContentDAO
	S3Servant           *storages3.Servant
	MaxUploadFileSize   int64
}
