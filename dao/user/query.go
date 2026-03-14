package user

import (
	"miaoverse/model/dao/user"
)

// QueryByPhone 按手机号查询所有用户（适配一个手机号多账户）
func (d *UserDAO) QueryByPhone(phone string) ([]user.User, error) {
	var users []user.User
	err := d.DB.Where("phone = ?", phone).Find(&users).Error
	return users, err
}

// QueryCredential 按用户ID+凭证类型查询凭证
func (d *UserDAO) QueryCredential(userID uint64, credType int8) (*user.UserCredential, error) {
	var cred user.UserCredential
	err := d.DB.Where("user_id = ? AND credential_type = ?", userID, credType).First(&cred).Error
	return &cred, err
}
