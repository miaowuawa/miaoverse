package interacts

import (
	"gorm.io/gorm"
	"miaoverse/consts"
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
		Update("status", consts.CommentStatusDeleted).Error
}

// CreateCommentAndMeta 创建评论并原子自增动态评论计数（事务），同时写入互动记录（type=comment，target_type=moment）。
func (d *InteractsDAO) CreateCommentAndMeta(comment modelinteracts.Comment, momentID uint64, momentAuthor uint32) (*modelinteracts.Comment, error) {
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		interact := modelinteracts.Interacts{
			UserFrom:   comment.UserID,
			UserTo:     momentAuthor,
			TargetID:   momentID,
			Type:       consts.InteractTypeComment,
			TargetType: consts.InteractTargetMoment,
			Status:     consts.InteractStatusNormal,
		}
		if err := tx.Create(&interact).Error; err != nil {
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

// CreateReplyCommentAndInteract 创建楼中楼回复并写入互动记录（事务，原子性）。
// 互动记录：type=reply，target_type=comment，user_to=被回复评论作者，reference_id=楼中楼首条评论 id。
// 楼中楼回复不计入 moment_meta.comment_count（该计数只统计 target_type=moment 的一级评论，见 CountMomentCommentsReal）。
func (d *InteractsDAO) CreateReplyCommentAndInteract(comment modelinteracts.Comment, interact modelinteracts.Interacts) (*modelinteracts.Comment, error) {
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		return tx.Create(&interact).Error
	})
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// QueryCommentRepliesByRoot 查询楼中楼完整回复列表：以 rootID 为根，用递归 CTE（WITH RECURSIVE）
// 在数据库内部做迭代式 BFS，单条 SQL 一次往返收集全部子孙回复。
// 相比逐层循环：无多次往返、无逐层膨胀的超大 IN 列表，且每行扩展都走 idx_comment_target 索引。
// 返回按 id 升序的扁平列表（仅 status=normal）；maxDepth 限制最大嵌套层数，防止脏数据成环导致无限递归。
// 要求 MySQL 8.0+ / MariaDB 10.2+（WITH RECURSIVE）。
func (d *InteractsDAO) QueryCommentRepliesByRoot(rootID uint64, maxDepth int) ([]modelinteracts.Comment, error) {
	const sql = `
WITH RECURSIVE reply_tree AS (
    SELECT id, user_id, target_id, target_type, content, status, created_at, updated_at, 0 AS depth
    FROM comment
    WHERE id = ? AND status = ?
    UNION ALL
    SELECT c.id, c.user_id, c.target_id, c.target_type, c.content, c.status, c.created_at, c.updated_at, rt.depth + 1
    FROM comment c
    INNER JOIN reply_tree rt ON c.target_id = rt.id
    WHERE c.target_type = ? AND c.status = ? AND rt.depth < ?
)
SELECT id, user_id, target_id, target_type, content, status, created_at, updated_at
FROM reply_tree
WHERE id <> ?
ORDER BY id ASC`

	var list []modelinteracts.Comment
	err := d.DB.Raw(sql,
		rootID,
		consts.CommentStatusNormal,
		consts.CommentTargetComment,
		consts.CommentStatusNormal,
		maxDepth,
		rootID,
	).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CountMomentCommentsReal 统计动态实际评论数（target_type=moment, status=normal）。
// 楼中楼回复（target_type=comment）不计入动态评论数。
func (d *InteractsDAO) CountMomentCommentsReal(momentID uint64) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Comment{}).
		Where("target_id = ? AND target_type = ? AND status = ?",
			momentID, consts.CommentTargetMoment, consts.CommentStatusNormal).
		Count(&count).Error
	return count, err
}

// CountMomentCommentsBatch 批量统计多个动态的实际评论数（一次 GROUP BY 查询）。
func (d *InteractsDAO) CountMomentCommentsBatch(momentIDs []uint64) (map[uint64]int64, error) {
	result := map[uint64]int64{}
	if len(momentIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		TargetID uint64
		Count    int64
	}
	err := d.DB.Model(&modelinteracts.Comment{}).
		Select("target_id, COUNT(*) AS count").
		Where("target_id IN ? AND target_type = ? AND status = ?",
			momentIDs, consts.CommentTargetMoment, consts.CommentStatusNormal).
		Group("target_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.TargetID] = row.Count
	}
	return result, nil
}
