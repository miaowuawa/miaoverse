package user

import (
	"miaoverse/model/dao/user"
)

func (d *UserDAO) Create(user *user.User, credential *user.UserCredential) (uint64, error) {

	// 开启事务
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 创建用户
	if err := d.DB.Create(user).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// 2. 创建凭证（关联用户ID）
	credential.UserID = user.ID
	if err := tx.Create(credential).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return user.ID, nil
}
