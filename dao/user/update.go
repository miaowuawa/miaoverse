package user

import modeluser "miaoverse/model/dao/user"

func (d *UserDAO) UpdateProfile(userID uint64, updates map[string]any) (*modeluser.User, error) {
	if err := d.DB.Model(&modeluser.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return d.QueryByID(userID)
}
