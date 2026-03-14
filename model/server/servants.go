package server

import (
	"miaoverse/service/security/sms/codemanager"
	"miaoverse/service/security/sms/smsbao"
)

type Servants struct {
	SmsServant  *smsbao.SmsBaoServant
	CodeManager *codemanager.CodeManager
}
