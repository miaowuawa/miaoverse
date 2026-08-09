package content

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"miaoverse/consts"
	"miaoverse/model/dao/moment"
)

func (d *ContentDAO) CreateMoment(m moment.Moment) (*moment.Moment, error) {
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		// 同步创建计数元数据，保证后续计数自增有目标行
		return tx.Create(&moment.MomentInteractCount{MomentID: m.ID}).Error
	})
	if err != nil {
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

// CountVisibleMomentsByUser 统计目标用户对查看者可见的动态数量。
// 可见规则：公开（0）全部可见；仅好友（1）需查看者关注目标；仅粉丝（3）需目标关注查看者；仅自己（2）仅本人可见。
func (d *ContentDAO) CountVisibleMomentsByUser(viewerID uint32, targetUserID uint32, isFriend bool, isFan bool) (int64, error) {
	var count int64
	err := d.DB.Model(&moment.Moment{}).
		Where("user_id = ? AND status = ?", targetUserID, consts.MomentStatusNormal).
		Where(d.visibleMomentScope(viewerID, targetUserID, isFriend, isFan)).
		Count(&count).Error
	return count, err
}

// QueryVisibleMomentsByUser 分页查询目标用户对查看者可见的动态列表（含互动计数）。
func (d *ContentDAO) QueryVisibleMomentsByUser(viewerID uint32, targetUserID uint32, isFriend bool, isFan bool, offset, limit int) ([]moment.Moment, error) {
	var list []moment.Moment
	err := d.DB.Model(&moment.Moment{}).
		Where("user_id = ? AND status = ?", targetUserID, consts.MomentStatusNormal).
		Where(d.visibleMomentScope(viewerID, targetUserID, isFriend, isFan)).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// visibleMomentScope 构造动态可见性 SQL 条件
func (d *ContentDAO) visibleMomentScope(viewerID uint32, targetUserID uint32, isFriend bool, isFan bool) *gorm.DB {
	scope := d.DB.Where("permission = ?", consts.MomentPermissionPublic)
	if viewerID == targetUserID {
		scope = scope.Or("permission = ?", consts.MomentPermissionPrivate)
	}
	if isFriend {
		scope = scope.Or("permission = ?", consts.MomentPermissionFriends)
	}
	if isFan {
		scope = scope.Or("permission = ?", consts.MomentPermissionFans)
	}
	return scope
}

func (d *ContentDAO) UpdateMoment(id uint64, updates map[string]any) error {
	return d.DB.Model(&moment.Moment{}).Where("id = ?", id).Updates(updates).Error
}

func (d *ContentDAO) DeleteMoment(id uint64) error {
	return d.DB.Model(&moment.Moment{}).Where("id = ?", id).
		Update("status", consts.MomentStatusDeleted).Error
}

func (d *ContentDAO) CreateMomentInteractCount(meta moment.MomentInteractCount) error {
	return d.DB.Create(&meta).Error
}

func (d *ContentDAO) QueryMomentInteractCount(momentID uint64) (*moment.MomentInteractCount, error) {
	var meta moment.MomentInteractCount
	err := d.DB.Where("moment_id = ?", momentID).First(&meta).Error
	return &meta, err
}

func (d *ContentDAO) IncrementMomentInteractCount(momentID uint64, field string, delta int64) error {
	return d.DB.Model(&moment.MomentInteractCount{}).
		Where("moment_id = ?", momentID).
		UpdateColumn(field, gorm.Expr(field+" + ?", delta)).Error
}

// QueryMomentInteractCounts 批量查询动态互动计数
func (d *ContentDAO) QueryMomentInteractCounts(momentIDs []uint64) (map[uint64]moment.MomentInteractCount, error) {
	result := map[uint64]moment.MomentInteractCount{}
	if len(momentIDs) == 0 {
		return result, nil
	}

	var list []moment.MomentInteractCount
	if err := d.DB.Where("moment_id IN ?", momentIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, meta := range list {
		result[meta.MomentID] = meta
	}
	return result, nil
}

// QueryAllMomentIDs 查询全部动态 ID（用于定期计数校准）
func (d *ContentDAO) QueryAllMomentIDs() ([]moment.Moment, error) {
	var list []moment.Moment
	err := d.DB.Select("id").Find(&list).Error
	return list, err
}

// QueryRecentMomentIDs 查询最近 N 分钟内更新过的动态 ID（用于增量校准）。
// 动态本身、评论、点赞都会刷新 moment.updated_at 或 moment_interact_count.updated_at。
func (d *ContentDAO) QueryRecentMomentIDs(since time.Time) ([]uint64, error) {
	var ids []uint64
	err := d.DB.Model(&moment.Moment{}).
		Select("id").
		Where("updated_at >= ?", since).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// SetMomentInteractCounts 批量覆盖动态计数（用于定期校准，保证计数与互动记录实际数量同步）
func (d *ContentDAO) SetMomentInteractCounts(updates map[uint64]moment.MomentInteractCount) error {
	for momentID, meta := range updates {
		if err := d.DB.Model(&moment.MomentInteractCount{}).
			Where("moment_id = ?", momentID).
			Updates(map[string]any{
				"like_count":    meta.LikeCount,
				"comment_count": meta.CommentCount,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

// UpsertMomentInteractCounts 批量覆盖动态计数（单条 SQL 批量 upsert，避免逐条 UPDATE）。
func (d *ContentDAO) UpsertMomentInteractCounts(updates map[uint64]moment.MomentInteractCount) error {
	if len(updates) == 0 {
		return nil
	}

	rows := make([]moment.MomentInteractCount, 0, len(updates))
	for _, meta := range updates {
		rows = append(rows, meta)
	}
	return d.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "moment_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"like_count", "comment_count"}),
	}).Create(&rows).Error
}
