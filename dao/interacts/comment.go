package interacts

import (
	modelinteracts "miaoverse/model/dao/interacts"
)

func (d *InteractsDAO) CreateComment(c modelinteracts.Comment) (*modelinteracts.Comment, error) {
	if err := d.DB.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (d *InteractsDAO) QueryCommentByID(id uint64) (*modelinteracts.Comment, error) {
	var c modelinteracts.Comment
	err := d.DB.Where("id = ?", id).First(&c).Error
	return &c, err
}

func (d *InteractsDAO) QueryCommentsByTarget(targetID uint64, targetType uint8, offset, limit int) ([]modelinteracts.Comment, error) {
	var list []modelinteracts.Comment
	err := d.DB.Where("target_id = ? AND target_type = ?", targetID, targetType).
		Order("id ASC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

func (d *InteractsDAO) DeleteComment(id uint64) error {
	return d.DB.Model(&modelinteracts.Comment{}).Where("id = ?", id).
		Update("status", modelinteracts.CommentStatusDeleted).Error
}
