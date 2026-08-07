package content

import (
	"gorm.io/gorm"
	"miaoverse/model/dao/moment"
)

func (d *ContentDAO) CreateMoment(m moment.Moment) (*moment.Moment, error) {
	if err := d.DB.Create(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (d *ContentDAO) QueryMomentByID(id uint64) (*moment.Moment, error) {
	var m moment.Moment
	err := d.DB.Where("id = ?", id).First(&m).Error
	return &m, err
}

func (d *ContentDAO) QueryMomentsByUser(userID uint32, offset, limit int) ([]moment.Moment, error) {
	var list []moment.Moment
	err := d.DB.Where("user_id = ?", userID).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

func (d *ContentDAO) UpdateMoment(id uint64, updates map[string]any) error {
	return d.DB.Model(&moment.Moment{}).Where("id = ?", id).Updates(updates).Error
}

func (d *ContentDAO) DeleteMoment(id uint64) error {
	return d.DB.Model(&moment.Moment{}).Where("id = ?", id).
		Update("status", moment.MomentStatusDeleted).Error
}

func (d *ContentDAO) CreateMomentMeta(meta moment.MomentMetaData) error {
	return d.DB.Create(&meta).Error
}

func (d *ContentDAO) QueryMomentMeta(momentID uint64) (*moment.MomentMetaData, error) {
	var meta moment.MomentMetaData
	err := d.DB.Where("moment_id = ?", momentID).First(&meta).Error
	return &meta, err
}

func (d *ContentDAO) IncrementMomentMeta(momentID uint64, field string, delta int64) error {
	return d.DB.Model(&moment.MomentMetaData{}).
		Where("moment_id = ?", momentID).
		UpdateColumn(field, gorm.Expr(field+" + ?", delta)).Error
}
