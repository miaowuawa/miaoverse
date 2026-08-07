package user

import (
	modeluser "miaoverse/model/dao/user"
)

func (d *UserDAO) CreateNotify(n modeluser.Notify) error {
	return d.DB.Create(&n).Error
}

func (d *UserDAO) QueryNotifiesByUser(userID uint32, offset, limit int) ([]modeluser.Notify, error) {
	var list []modeluser.Notify
	err := d.DB.Where("user_id = ?", userID).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

func (d *UserDAO) MarkNotifyRead(id uint64, userID uint32) error {
	return d.DB.Model(&modeluser.Notify{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", modeluser.NotifyStatusRead).Error
}

func (d *UserDAO) MarkAllNotifyRead(userID uint32) error {
	return d.DB.Model(&modeluser.Notify{}).
		Where("user_id = ? AND status = ?", userID, modeluser.NotifyStatusUnread).
		Update("status", modeluser.NotifyStatusRead).Error
}

func (d *UserDAO) DeleteNotify(id uint64, userID uint32) error {
	return d.DB.Model(&modeluser.Notify{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", modeluser.NotifyStatusDeleted).Error
}

func (d *UserDAO) CountUnreadNotifies(userID uint32) (int64, error) {
	var count int64
	err := d.DB.Model(&modeluser.Notify{}).
		Where("user_id = ? AND status = ?", userID, modeluser.NotifyStatusUnread).
		Count(&count).Error
	return count, err
}
