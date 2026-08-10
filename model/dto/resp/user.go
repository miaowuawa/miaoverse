package resp

import "miaoverse/model/dao/user"

type CodeWithMsgUserID struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	UID  uint32 `json:"uid"`
}

// CodeWithMsgAccountList 当前会话手机号绑定的账号列表。
type CodeWithMsgAccountList struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Current uint32      `json:"current"`
	Users   []user.User `json:"users"`
}
