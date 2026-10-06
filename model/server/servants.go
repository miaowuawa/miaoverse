package server

import (
	"github.com/go-playground/validator/v10"
	fiberstoreredis "github.com/gofiber/storage/redis/v3"
	"miaoverse/dao/article"
	"miaoverse/dao/content"
	"miaoverse/dao/interacts"
	"miaoverse/dao/sticker"
	"miaoverse/dao/user"
	"miaoverse/service/Notify"
	"miaoverse/service/UserBlock"
	storagemongo "miaoverse/service/mongo"
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
	InteractsServant    *interacts.InteractsDAO
	ArticleServant      *article.ArticleDAO
	StickerServant      *sticker.StickerDAO
	BlockServant        *UserBlock.Servant
	NotifyServant       *Notify.Servant
	S3Servant           *storages3.Servant
	MongoServant        *storagemongo.Servant
	MaxUploadFileSize   int64
	MaxStickerFileSize  int64
}
