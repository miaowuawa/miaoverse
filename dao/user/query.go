package user

import (
	"strconv"

	"miaoverse/consts"
	"miaoverse/model/dao/user"
)

func (d *UserDAO) QueryByPhone(phone string, region uint16) ([]user.User, error) {
	var users []user.User
	err := d.DB.Table("`user`").
		Select("`user`.*").
		Joins("JOIN user_credentials AS phone_cred ON phone_cred.user_id = `user`.id").
		Joins("JOIN user_credentials AS region_cred ON region_cred.user_id = `user`.id").
		Where("phone_cred.credential_type = ? AND phone_cred.credential_key = ? AND phone_cred.credential_value = ?", consts.Phone, "phone", phone).
		Where("region_cred.credential_type = ? AND region_cred.credential_key = ? AND region_cred.credential_value = ?", consts.Phone, "region", strconv.FormatUint(uint64(region), 10)).
		Find(&users).Error
	return users, err
}

func (d *UserDAO) QueryByID(userID uint32) (*user.User, error) {
	var u user.User
	err := d.DB.Where("id = ?", userID).First(&u).Error
	return &u, err
}

func (d *UserDAO) QueryCredential(userID uint32, credType uint8) (*user.UserCredential, error) {
	var cred user.UserCredential
	err := d.DB.Where("user_id = ? AND credential_type = ?", userID, credType).First(&cred).Error
	return &cred, err
}

func (d *UserDAO) HasCredential(userID uint32, credType uint8) (bool, error) {
	var count int64
	err := d.DB.Model(&user.UserCredential{}).
		Where("user_id = ? AND credential_type = ?", userID, credType).
		Count(&count).Error
	return count > 0, err
}
