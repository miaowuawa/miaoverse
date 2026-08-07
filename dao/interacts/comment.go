package interacts

import (
	"gorm.io/gorm"
	modelinteracts "miaoverse/model/dao/interacts"
	modelmoment "miaoverse/model/dao/moment"
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

// CreateCommentAndMeta 创建评论并原子自增动态评论计数（事务）。
func (d *InteractsDAO) CreateCommentAndMeta(comment modelinteracts.Comment, momentID uint64) (*modelinteracts.Comment, error) {
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		return tx.Model(&modelmoment.MomentMetaData{}).
			Where("moment_id = ?", momentID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	})
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// CountMomentCommentsReal 统计动态实际评论数（target_type=moment, status=normal）。
// 楼中楼回复（target_type=comment）不计入动态评论数。
func (d *InteractsDAO) CountMomentCommentsReal(momentID uint64) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Comment{}).
		Where("target_id = ? AND target_type = ? AND status = ?",
			momentID, modelinteracts.CommentTargetMoment, modelinteracts.CommentStatusNormal).
		Count(&count).Error
	return count, err
}
