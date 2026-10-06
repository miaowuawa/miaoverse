package smsreq

// GetSmsReq 申请短信验证码。
// Action 为业务场景（见 consts.Action*），不传时按登录/注册处理；场景参与验证码存储 key，
// 不同场景的验证码互不通用。
type GetSmsReq struct {
	Phone     string `json:"phone" validate:"required,isPhone"`
	Timestamp string `json:"a" validate:"required,isDigit"`
	Region    string `json:"region" validate:"required,numeric"`
	Action    string `json:"action"`
}
