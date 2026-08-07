package interacts

import (
	modelinteracts "miaoverse/model/dao/interacts"
)

func (d *InteractsDAO) CreateInteract(i modelinteracts.Interacts) error {
	return d.DB.Create(&i).Error
}

func (d *InteractsDAO) QueryInteract(userFrom uint32, targetID uint64, interactType uint8) (*modelinteracts.Interacts, error) {
	var i modelinteracts.Interacts
	err := d.DB.Where("user_from = ? AND target_id = ? AND type = ?", userFrom, targetID, interactType).
		Order("id DESC").First(&i).Error
	return &i, err
}

func (d *InteractsDAO) RevokeInteract(id uint64, forced bool) error {
	status := modelinteracts.InteractStatusRevoked
	if forced {
		status = modelinteracts.InteractStatusForced
	}
	return d.DB.Model(&modelinteracts.Interacts{}).Where("id = ?", id).
		Update("status", status).Error
}

func (d *InteractsDAO) CountInteracts(targetID uint64, interactType uint8) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("target_id = ? AND type = ? AND status = ?", targetID, interactType, modelinteracts.InteractStatusNormal).
		Count(&count).Error
	return count, err
}

func (d *InteractsDAO) QueryInteractsByUser(userFrom uint32, offset, limit int) ([]modelinteracts.Interacts, error) {
	var list []modelinteracts.Interacts
	err := d.DB.Where("user_from = ?", userFrom).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// IsFollowing 查询 userFrom 是否正在关注 userTo（status 为正常）
func (d *InteractsDAO) IsFollowing(userFrom uint32, userTo uint32) (bool, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_from = ? AND user_to = ? AND type = ? AND status = ?",
			userFrom, userTo, modelinteracts.InteractTypeFollow, modelinteracts.InteractStatusNormal).
		Count(&count).Error
	return count > 0, err
}

// CountFollowing 统计 userID 正在关注的用户数量
func (d *InteractsDAO) CountFollowing(userID uint32) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_from = ? AND type = ? AND status = ?",
			userID, modelinteracts.InteractTypeFollow, modelinteracts.InteractStatusNormal).
		Count(&count).Error
	return count, err
}

// CountFollowers 统计关注 userID 的用户数量
func (d *InteractsDAO) CountFollowers(userID uint32) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_to = ? AND type = ? AND status = ?",
			userID, modelinteracts.InteractTypeFollow, modelinteracts.InteractStatusNormal).
		Count(&count).Error
	return count, err
}

// QueryFollowingByUser 分页查询 userID 正在关注的用户列表（user_from = userID）
func (d *InteractsDAO) QueryFollowingByUser(userID uint32, offset, limit int) ([]modelinteracts.Interacts, error) {
	var list []modelinteracts.Interacts
	err := d.DB.Where("user_from = ? AND type = ? AND status = ?",
		userID, modelinteracts.InteractTypeFollow, modelinteracts.InteractStatusNormal).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// QueryFollowersByUser 分页查询关注 userID 的用户列表（user_to = userID）
func (d *InteractsDAO) QueryFollowersByUser(userID uint32, offset, limit int) ([]modelinteracts.Interacts, error) {
	var list []modelinteracts.Interacts
	err := d.DB.Where("user_to = ? AND type = ? AND status = ?",
		userID, modelinteracts.InteractTypeFollow, modelinteracts.InteractStatusNormal).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// CountMomentLikesReal 统计动态实际点赞数（type=like, target_type=moment, status=normal）
func (d *InteractsDAO) CountMomentLikesReal(momentID uint64) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("target_id = ? AND type = ? AND target_type = ? AND status = ?",
			momentID, modelinteracts.InteractTypeLike, modelinteracts.InteractTargetMoment, modelinteracts.InteractStatusNormal).
		Count(&count).Error
	return count, err
}
