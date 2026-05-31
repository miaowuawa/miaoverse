package phone

import (
	"miaoverse/consts"
	"miaoverse/model/dao/user"
	"strconv"
)

func PhoneToCredentialStructure(phone string, region int) ([]user.UserCredential, error) {
	Credentials := []user.UserCredential{
		{
			CredentialType:  consts.Phone,
			CredentialKey:   "phone",
			CredentialValue: phone,
		},
		{
			CredentialType:  consts.Phone,
			CredentialKey:   "region",
			CredentialValue: strconv.Itoa(region),
		},
	}
	return Credentials, nil
}
