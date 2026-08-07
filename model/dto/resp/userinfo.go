package resp

import "miaoverse/model/dao/user"

type UserInfo struct {
	user.User
	BlockStatus    uint8  `json:"block_status"`
	PunishmentMask uint32 `json:"punishment_mask"`
}

type CodeWithMsgUserInfo struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	User UserInfo `json:"user"`
}
