package resp

import "miaoverse/model/dao/user"

type CodeWithMsgUserChoice struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Choices []user.User `json:"users"`
}
