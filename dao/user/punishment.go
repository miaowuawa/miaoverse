package user

import (
	"time"

	modeluser "miaoverse/model/dao/user"
)

// QueryPunishmentsByUser 查询用户全部惩罚记录（按封禁开始时间倒序）
func (d *UserDAO) QueryPunishmentsByUser(userID uint32) ([]modeluser.Punishment, error) {
	var list []modeluser.Punishment
	err := d.DB.Where("user_id = ?", userID).
		Order("punishment_time DESC, id DESC").
		Find(&list).Error
	return list, err
}

// HasActivePunishment 查询用户是否存在指定权限位的生效中惩罚记录。
// 生效中：punishment_status = active，且（无结束时间 或 结束时间晚于 now）。
func (d *UserDAO) HasActivePunishment(userID uint32, perm uint32, now time.Time) (bool, error) {
	var count int64
	err := d.DB.Model(&modeluser.Punishment{}).
		Where("user_id = ? AND punishment_status = ? AND (punishment_end_time IS NULL OR punishment_end_time > ?)",
			userID, modeluser.PunishmentStatusActive, now).
		Where("(punishment_type & ?) = ?", perm, perm).
		Count(&count).Error
	return count > 0, err
}

// QueryActivePunishmentMask 查询用户当前生效中的所有被封权限位（按位或合并），0 表示无生效封禁。
func (d *UserDAO) QueryActivePunishmentMask(userID uint32, now time.Time) (uint32, error) {
	var types []uint32
	err := d.DB.Model(&modeluser.Punishment{}).
		Where("user_id = ? AND punishment_status = ? AND (punishment_end_time IS NULL OR punishment_end_time > ?)",
			userID, modeluser.PunishmentStatusActive, now).
		Pluck("punishment_type", &types).Error
	if err != nil {
		return 0, err
	}

	var mask uint32
	for _, t := range types {
		mask |= t
	}
	return mask, nil
}
