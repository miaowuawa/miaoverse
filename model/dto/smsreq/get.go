package smsreq

type GetSmsReq struct {
	Phone     string `json:"phone" validate:"required,isDigit"`
	Timestamp string `json:"a" validate:"required,isDigit"`
	Region    string `json:"region" validate:"required,numeric"`
}
