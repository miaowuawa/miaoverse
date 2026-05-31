package smsresp

type SmsResp struct {
	CodeUUID string `json:"code_uuid"`
	Msg      string `json:"msg"`
}

type ValidateResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}
