package smsreq

type GetSmsReq struct {
	Phone     string `json:"phone"`
	Timestamp string `json:"a"`
	Region    string `json:"region"`
}
