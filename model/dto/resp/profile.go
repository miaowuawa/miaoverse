package resp

import "miaoverse/model/dao/user"

type CodeWithMsgUser struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	User user.User `json:"user"`
	// Phone 为打码后的绑定手机号（如 "+86 138****8000"），仅 GET /api/v1/user/me 返回，
	// 用于个人资料页只读展示。其他接口不填该字段（omitempty 不会出现在响应中）。
	// 手机号保存在 user_credentials 表，不属于 user 表字段，且当前不支持更改。
	Phone string `json:"phone,omitempty"`
}
