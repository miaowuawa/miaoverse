package smsbaoreq

// PhoneCaptchaRsp 短信宝响应结构体
type PhoneCaptchaRsp struct {
	Status string `json:"status"` // 响应状态，如 success 或 error
	Code   string `json:"code"`   // 状态码
	Msg    string `json:"msg"`    // 状态消息
}
