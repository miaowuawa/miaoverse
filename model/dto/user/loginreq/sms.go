package loginreq

type SMS struct {
	TimeStamp string `json:"a" validate:"required,isDigit"`
	Phone     string `json:"phone" validate:"required,isPhone"`
	Region    uint16 `json:"region" validate:"required,numeric"`
	UUID      string `json:"uuid" validate:"required,isUUIDv4"`
	Code      int    `json:"code" validate:"required,numeric"`
}

type ChooseAccount struct {
	UID uint32 `json:"uid" validate:"required"`
}

type SwitchAccount struct {
	UID uint32 `json:"uid" validate:"required"`
}
