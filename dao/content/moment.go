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

// CountVisibleMomentsByUser 统计目标用户对查看者可见的动态数量。
// 可见规则：公开（0）全部可见；仅好友（1）需查看者关注目标；仅粉丝（3）需目标关注查看者；仅自己（2）仅本人可见。
func (d *ContentDAO) CountVisibleMomentsByUser(viewerID uint32, targetUserID uint32, isFriend bool, isFan bool) (int64, error) {
	var count int64
	err := d.DB.Model(&moment.Moment{}).
		Where("user_id = ? AND status = ?", targetUserID, moment.MomentStatusNormal).
		Where(d.visibleMomentScope(viewerID, targetUserID, isFriend, isFan)).
		Count(&count).Error
	return count, err
}

// QueryVisibleMomentsByUser 分页查询目标用户对查看者可见的动态列表（含互动计数）。
func (d *ContentDAO) QueryVisibleMomentsByUser(viewerID uint32, targetUserID uint32, isFriend bool, isFan bool, offset, limit int) ([]moment.Moment, error) {
	var list []moment.Moment
	err := d.DB.Model(&moment.Moment{}).
		Where("user_id = ? AND status = ?", targetUserID, moment.MomentStatusNormal).
		Where(d.visibleMomentScope(viewerID, targetUserID, isFriend, isFan)).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// visibleMomentScope 构造动态可见性 SQL 条件
func (d *ContentDAO) visibleMomentScope(viewerID uint32, targetUserID uint32, isFriend bool, isFan bool) *gorm.DB {
	scope := d.DB.Where("permission = ?", moment.MomentPermissionPublic)
	if viewerID == targetUserID {
		scope = scope.Or("permission = ?", moment.MomentPermissionPrivate)
	}
	if isFriend {
		scope = scope.Or("permission = ?", moment.MomentPermissionFriends)
	}
	if isFan {
		scope = scope.Or("permission = ?", moment.MomentPermissionFans)
	}
	return scope
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

// QueryMomentMetas 批量查询动态互动计数
func (d *ContentDAO) QueryMomentMetas(momentIDs []uint64) (map[uint64]moment.MomentMetaData, error) {
	result := map[uint64]moment.MomentMetaData{}
	if len(momentIDs) == 0 {
		return result, nil
	}

	var list []moment.MomentMetaData
	if err := d.DB.Where("moment_id IN ?", momentIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, meta := range list {
		result[meta.MomentID] = meta
	}
	return result, nil
}
