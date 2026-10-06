package user

import (
	"time"

	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
)

// CreateNotify 写入通知记录（写入后回填自增 ID，供 SSE 推送与响应使用）。
func (d *UserDAO) CreateNotify(n *modeluser.Notify) error {
	return d.DB.Create(n).Error
}

// QueryNotifiesByUser 分页查询用户通知（id 倒序，排除已删除）。
// types 为通知类型白名单（如「账号」分类对应 0、1）；传 nil 表示不按类型过滤。
func (d *UserDAO) QueryNotifiesByUser(userID uint32, types []uint8, offset, limit int) ([]modeluser.Notify, error) {
	var list []modeluser.Notify
	q := d.DB.Where("user_id = ? AND status <> ?", userID, consts.NotifyStatusDeleted)
	if len(types) > 0 {
		q = q.Where("type IN ?", types)
	}
	err := q.Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// CountNotifiesByUser 统计用户通知总数（排除已删除），types 语义同 QueryNotifiesByUser。
func (d *UserDAO) CountNotifiesByUser(userID uint32, types []uint8) (int64, error) {
	var count int64
	q := d.DB.Model(&modeluser.Notify{}).
		Where("user_id = ? AND status <> ?", userID, consts.NotifyStatusDeleted)
	if len(types) > 0 {
		q = q.Where("type IN ?", types)
	}
	err := q.Count(&count).Error
	return count, err
}

// MarkNotifyRead 标记单条通知已读（幂等：仅未读时生效，同时记录 read_at）。
func (d *UserDAO) MarkNotifyRead(id uint64, userID uint32) error {
	now := time.Now()
	return d.DB.Model(&modeluser.Notify{}).
		Where("id = ? AND user_id = ? AND status = ?", id, userID, consts.NotifyStatusUnread).
		Updates(map[string]interface{}{
			"status":  consts.NotifyStatusRead,
			"read_at": now,
		}).Error
}

// MarkAllNotifyRead 将用户全部未读通知标记为已读（幂等）。
func (d *UserDAO) MarkAllNotifyRead(userID uint32) error {
	now := time.Now()
	return d.DB.Model(&modeluser.Notify{}).
		Where("user_id = ? AND status = ?", userID, consts.NotifyStatusUnread).
		Updates(map[string]interface{}{
			"status":  consts.NotifyStatusRead,
			"read_at": now,
		}).Error
}

// DeleteNotify 删除单条通知（软删除，幂等）。
func (d *UserDAO) DeleteNotify(id uint64, userID uint32) error {
	return d.DB.Model(&modeluser.Notify{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", consts.NotifyStatusDeleted).Error
}

// CountUnreadNotifies 统计用户全部未读通知数。
func (d *UserDAO) CountUnreadNotifies(userID uint32) (int64, error) {
	var count int64
	err := d.DB.Model(&modeluser.Notify{}).
		Where("user_id = ? AND status = ?", userID, consts.NotifyStatusUnread).
		Count(&count).Error
	return count, err
}

// CountUnreadNotifiesByTypes 按通知类型分组统计未读数（一次 GROUP BY，避免逐类型 COUNT）。
func (d *UserDAO) CountUnreadNotifiesByTypes(userID uint32) (map[uint8]int64, error) {
	result := map[uint8]int64{}
	var rows []struct {
		Type  uint8
		Count int64
	}
	err := d.DB.Model(&modeluser.Notify{}).
		Select("type, COUNT(*) AS count").
		Where("user_id = ? AND status = ?", userID, consts.NotifyStatusUnread).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.Type] = row.Count
	}
	return result, nil
}

// MarkNotifyReceived 标记通知已通过 SSE 推送给用户（received=1，仅记录送达状态，不影响已读状态）。
func (d *UserDAO) MarkNotifyReceived(id uint64) error {
	return d.DB.Model(&modeluser.Notify{}).
		Where("id = ?", id).
		Update("received", true).Error
}
