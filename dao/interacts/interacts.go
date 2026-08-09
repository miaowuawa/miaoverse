package interacts

import (
	"errors"

	"gorm.io/gorm"
	"miaoverse/consts"
	modelinteracts "miaoverse/model/dao/interacts"
	modelmoment "miaoverse/model/dao/moment"
)

func (d *InteractsDAO) CreateInteract(i modelinteracts.Interacts) error {
	return d.DB.Create(&i).Error
}

// FollowUser 关注用户（幂等：已存在正常关注时直接返回）。
func (d *InteractsDAO) FollowUser(userFrom uint32, userTo uint32) error {
	existing, err := d.QueryInteract(userFrom, uint64(userTo), consts.InteractTypeFollow)
	if err == nil && existing != nil && existing.Status == consts.InteractStatusNormal {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	interact := modelinteracts.Interacts{
		UserFrom:   userFrom,
		UserTo:     userTo,
		TargetID:   uint64(userTo),
		Type:       consts.InteractTypeFollow,
		TargetType: consts.InteractTargetUser,
		Status:     consts.InteractStatusNormal,
	}
	return d.DB.Create(&interact).Error
}

// UnfollowUser 取消关注（软撤销：将正常关注置为已撤销）。
func (d *InteractsDAO) UnfollowUser(userFrom uint32, userTo uint32) error {
	return d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_from = ? AND user_to = ? AND type = ? AND status = ?",
			userFrom, userTo, consts.InteractTypeFollow, consts.InteractStatusNormal).
		Update("status", consts.InteractStatusRevoked).Error
}

// LikeMomentAndMeta 点赞动态并原子自增动态点赞计数（事务，幂等）。
func (d *InteractsDAO) LikeMomentAndMeta(userID uint32, momentID uint64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&modelinteracts.Interacts{}).
			Where("user_from = ? AND target_id = ? AND type = ? AND target_type = ? AND status = ?",
				userID, momentID, consts.InteractTypeLike, consts.InteractTargetMoment, consts.InteractStatusNormal).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		interact := modelinteracts.Interacts{
			UserFrom:   userID,
			UserTo:     0,
			TargetID:   momentID,
			Type:       consts.InteractTypeLike,
			TargetType: consts.InteractTargetMoment,
			Status:     consts.InteractStatusNormal,
		}
		if err := tx.Create(&interact).Error; err != nil {
			return err
		}
		return tx.Model(&modelmoment.MomentInteractCount{}).
			Where("moment_id = ?", momentID).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
	})
}

// UnlikeMomentAndMeta 取消点赞并原子自减动态点赞计数（事务，幂等）。
func (d *InteractsDAO) UnlikeMomentAndMeta(userID uint32, momentID uint64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&modelinteracts.Interacts{}).
			Where("user_from = ? AND target_id = ? AND type = ? AND target_type = ? AND status = ?",
				userID, momentID, consts.InteractTypeLike, consts.InteractTargetMoment, consts.InteractStatusNormal).
			Update("status", consts.InteractStatusRevoked)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&modelmoment.MomentInteractCount{}).
			Where("moment_id = ?", momentID).
			UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error
	})
}

func (d *InteractsDAO) QueryInteract(userFrom uint32, targetID uint64, interactType uint8) (*modelinteracts.Interacts, error) {
	var i modelinteracts.Interacts
	err := d.DB.Where("user_from = ? AND target_id = ? AND type = ?", userFrom, targetID, interactType).
		Order("id DESC").First(&i).Error
	return &i, err
}

func (d *InteractsDAO) RevokeInteract(id uint64, forced bool) error {
	status := consts.InteractStatusRevoked
	if forced {
		status = consts.InteractStatusForced
	}
	return d.DB.Model(&modelinteracts.Interacts{}).Where("id = ?", id).
		Update("status", status).Error
}

func (d *InteractsDAO) CountInteracts(targetID uint64, interactType uint8) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("target_id = ? AND type = ? AND status = ?", targetID, interactType, consts.InteractStatusNormal).
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
			userFrom, userTo, consts.InteractTypeFollow, consts.InteractStatusNormal).
		Count(&count).Error
	return count > 0, err
}

// CountFollowing 统计 userID 正在关注的用户数量
func (d *InteractsDAO) CountFollowing(userID uint32) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_from = ? AND type = ? AND status = ?",
			userID, consts.InteractTypeFollow, consts.InteractStatusNormal).
		Count(&count).Error
	return count, err
}

// CountFollowers 统计关注 userID 的用户数量
func (d *InteractsDAO) CountFollowers(userID uint32) (int64, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_to = ? AND type = ? AND status = ?",
			userID, consts.InteractTypeFollow, consts.InteractStatusNormal).
		Count(&count).Error
	return count, err
}

// QueryFollowingByUser 分页查询 userID 正在关注的用户列表（user_from = userID）
func (d *InteractsDAO) QueryFollowingByUser(userID uint32, offset, limit int) ([]modelinteracts.Interacts, error) {
	var list []modelinteracts.Interacts
	err := d.DB.Where("user_from = ? AND type = ? AND status = ?",
		userID, consts.InteractTypeFollow, consts.InteractStatusNormal).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// QueryFollowersByUser 分页查询关注 userID 的用户列表（user_to = userID）
func (d *InteractsDAO) QueryFollowersByUser(userID uint32, offset, limit int) ([]modelinteracts.Interacts, error) {
	var list []modelinteracts.Interacts
	err := d.DB.Where("user_to = ? AND type = ? AND status = ?",
		userID, consts.InteractTypeFollow, consts.InteractStatusNormal).
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
			momentID, consts.InteractTypeLike, consts.InteractTargetMoment, consts.InteractStatusNormal).
		Count(&count).Error
	return count, err
}

// CountMomentLikesBatch 批量统计多个动态的实际点赞数（一次 GROUP BY 查询）。
func (d *InteractsDAO) CountMomentLikesBatch(momentIDs []uint64) (map[uint64]int64, error) {
	result := map[uint64]int64{}
	if len(momentIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		TargetID uint64
		Count    int64
	}
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Select("target_id, COUNT(*) AS count").
		Where("target_id IN ? AND type = ? AND target_type = ? AND status = ?",
			momentIDs, consts.InteractTypeLike, consts.InteractTargetMoment, consts.InteractStatusNormal).
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
