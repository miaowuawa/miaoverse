package smsresp

type SmsResp struct {
	CodeId   string `json:"code"`
	CodeUuid string `json:"code_uuid"`
	Msg      string `json:"msg"`
}
