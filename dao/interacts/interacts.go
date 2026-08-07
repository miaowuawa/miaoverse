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
