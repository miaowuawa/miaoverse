package loginreq

type SMS struct {
	TimeStamp string `json:"a" validate:"required,isDigit"`
	Phone     string `json:"phone" validate:"required,isDigit"`
	Region    int    `json:"region" validate:"required,numeric"`
	UUID      string `json:"uuid" validate:"required,isUUIDv4"`
	Code      int    `json:"code" validate:"required,numeric"`
}

type ChooseAccount struct {
	UID uint64 `json:"uid" validate:"required"`
}
