package resp

import "miaoverse/model/dao/user"

type CodeWithMsgUser struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	User user.User `json:"user"`
}
