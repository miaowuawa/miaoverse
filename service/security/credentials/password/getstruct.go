package password

import (
	"miaoverse/consts"
	"miaoverse/model/dao/user"
	"miaoverse/util"
)

func PasswordToCredentialStructure(UserID uint32, password string) (error, []*user.UserCredential) {
	// 1. 对密码进行bcrypt加密
	bcryptHash, err := util.Security.HashPassword(password)
	if err != nil {
		return err, nil // 加密失败时返回错误和空数组
	}

	// 2. 创建单个凭证结构体
	credential := &user.UserCredential{
		UserID:          UserID,
		CredentialType:  consts.Password,
		CredentialKey:   "bcrypt",
		CredentialValue: bcryptHash,
	}

	// 3. 将单个凭证放入数组中返回
	credentials := []*user.UserCredential{credential}

	return nil, credentials
}
