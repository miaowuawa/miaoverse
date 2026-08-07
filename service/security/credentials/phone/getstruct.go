package phone

import (
	"strconv"

	"miaoverse/consts"
	"miaoverse/model/dao/user"
)

func PhoneToCredentialStructure(phone string, region uint16) ([]user.UserCredential, error) {
	Credentials := []user.UserCredential{
		{
			CredentialType:  consts.Phone,
			CredentialKey:   "phone",
			CredentialValue: phone,
		},
		{
			CredentialType:  consts.Phone,
			CredentialKey:   "region",
			CredentialValue: strconv.FormatUint(uint64(region), 10),
		},
	}
	return Credentials, nil
}
