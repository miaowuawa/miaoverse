package phone

import (
	"miaoverse/consts"
	"miaoverse/model/dao/user"
	"miaoverse/util/encrypt/bcrypthash"
)

func PhoneToCredentialStructure(UserID uint64, password string) (error, []*user.UserCredential) {
	// 1. 对密码进行bcrypt加密
	bcryptHash, err := bcrypthash.HashPassword(password)
	if err != nil {
		return err, nil // 加密失败时返回错误和空数组
	}

	// 2. 创建单个凭证结构体
	credential := &user.UserCredential{
		UserID:          UserID,
		CredentialType:  consts.Phone,
		CredentialKey:   "bcrypt",
		CredentialValue: bcryptHash,
	}

	// 3. 将单个凭证放入数组中返回
	credentials := []*user.UserCredential{credential}

	return nil, credentials
}
