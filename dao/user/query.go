package user

import (
	"strconv"

	"miaoverse/consts"
	"miaoverse/model/dao/user"
)

// QueryByPhone 查询指定手机号绑定的全部账号（含已注销），用于注册等存在性判断。
func (d *UserDAO) QueryByPhone(phone string, region uint16) ([]user.User, error) {
	return d.queryByPhone(phone, region, false)
}

// QueryByPhoneLoginable 查询指定手机号绑定的可登录账号。
// 已注销账号（UserStatusClosed）无法再登录，统一过滤，避免出现在登录选择、账号切换等列表中。
func (d *UserDAO) QueryByPhoneLoginable(phone string, region uint16) ([]user.User, error) {
	return d.queryByPhone(phone, region, true)
}

func (d *UserDAO) queryByPhone(phone string, region uint16, excludeClosed bool) ([]user.User, error) {
	var users []user.User
	query := d.DB.Table("`user`").
		Select("`user`.*").
		Joins("JOIN user_credentials AS phone_cred ON phone_cred.user_id = `user`.id").
		Joins("JOIN user_credentials AS region_cred ON region_cred.user_id = `user`.id").
		Where("phone_cred.credential_type = ? AND phone_cred.credential_key = ? AND phone_cred.credential_value = ?", consts.Phone, "phone", phone).
		Where("region_cred.credential_type = ? AND region_cred.credential_key = ? AND region_cred.credential_value = ?", consts.Phone, "region", strconv.FormatUint(uint64(region), 10))
	if excludeClosed {
		query = query.Where("`user`.status != ?", consts.UserStatusClosed)
	}
	err := query.Find(&users).Error
	return users, err
}

func (d *UserDAO) QueryByID(userID uint32) (*user.User, error) {
	var u user.User
	err := d.DB.Where("id = ?", userID).First(&u).Error
	return &u, err
}

// QueryUsersByIDs 批量按 ID 查询用户
func (d *UserDAO) QueryUsersByIDs(ids []uint32) (map[uint32]user.User, error) {
	result := map[uint32]user.User{}
	if len(ids) == 0 {
		return result, nil
	}

	var list []user.User
	if err := d.DB.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, u := range list {
		result[u.ID] = u
	}
	return result, nil
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
