package resp

import "miaoverse/model/dao/user"

type UserInfo struct {
	user.User
	BlockStatus    uint8  `json:"block_status"`
	PunishmentMask uint32 `json:"punishment_mask"`
	// Phone 为打码后的绑定手机号（如 "+86 138****8000"）。
	// 仅在「请求者就是被查询用户本人」时返回，查询他人资料时省略（omitempty），
	// 展示与资料接口一致：手机号不支持更改，只做只读展示。
	Phone string `json:"phone,omitempty"`
}

type CodeWithMsgUserInfo struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	User UserInfo `json:"user"`
}
