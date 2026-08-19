package interacts

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"miaoverse/consts"
	modelinteracts "miaoverse/model/dao/interacts"
	modelmoment "miaoverse/model/dao/moment"
)

func (d *InteractsDAO) CreateInteract(i modelinteracts.Interacts) error {
	return d.DB.Create(&i).Error
}

// FollowUser 关注用户（幂等，并发安全）。
// 依赖 interacts.single_key 唯一索引：重复关注时 INSERT 被数据库原子忽略（RowsAffected=0），
// 只把已撤销（status≠normal）的旧关注恢复为正常，不存在"先查再插"的竞态窗口。
func (d *InteractsDAO) FollowUser(userFrom uint32, userTo uint32) error {
	interact := modelinteracts.Interacts{
		UserFrom:   userFrom,
		UserTo:     userTo,
		TargetID:   uint64(userTo),
		Type:       consts.InteractTypeFollow,
		TargetType: consts.InteractTargetUser,
		Status:     consts.InteractStatusNormal,
	}
	created := d.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&interact)
	if created.Error != nil {
		return created.Error
	}
	if created.RowsAffected == 0 {
		// 已有互动行：仅当之前不是正常状态时恢复为正常（重复关注幂等）
		return d.DB.Model(&modelinteracts.Interacts{}).
			Where("user_from = ? AND target_id = ? AND type = ? AND target_type = ? AND status <> ?",
				userFrom, uint64(userTo), consts.InteractTypeFollow, consts.InteractTargetUser, consts.InteractStatusNormal).
			Updates(map[string]interface{}{
				"status":   consts.InteractStatusNormal,
				"acted_at": time.Now(),
			}).Error
	}
	return nil
}

// UnfollowUser 取消关注（软撤销：将正常关注置为已撤销）。
func (d *InteractsDAO) UnfollowUser(userFrom uint32, userTo uint32) error {
	return d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_from = ? AND user_to = ? AND type = ? AND status = ?",
			userFrom, userTo, consts.InteractTypeFollow, consts.InteractStatusNormal).
		Update("status", consts.InteractStatusRevoked).Error
}

// LikeMomentAndMeta 点赞动态并原子自增动态点赞计数（事务，幂等，并发安全）。
// momentAuthor 为动态作者 ID：interacts.user_to 有外键约束，必须指向真实用户（不能为 0）。
// 并发安全依赖 interacts.single_key 唯一索引：重复点赞时 INSERT 被数据库原子忽略（RowsAffected=0），
// 只把已撤销（status≠normal）的旧点赞恢复为正常，不存在"先查再插"的竞态窗口；
// 计数仅在真正新增或恢复为正常时 +1，重复点赞不重复计数。
func (d *InteractsDAO) LikeMomentAndMeta(userID uint32, momentID uint64, momentAuthor uint32) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		interact := modelinteracts.Interacts{
			UserFrom:   userID,
			UserTo:     momentAuthor,
			TargetID:   momentID,
			Type:       consts.InteractTypeLike,
			TargetType: consts.InteractTargetMoment,
			Status:     consts.InteractStatusNormal,
		}
		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&interact)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected == 0 {
			// 已有互动行：仅当之前不是正常状态时恢复为正常
			reactivated := tx.Model(&modelinteracts.Interacts{}).
				Where("user_from = ? AND target_id = ? AND type = ? AND target_type = ? AND status <> ?",
					userID, momentID, consts.InteractTypeLike, consts.InteractTargetMoment, consts.InteractStatusNormal).
				Updates(map[string]interface{}{
					"status":   consts.InteractStatusNormal,
					"acted_at": time.Now(),
				})
			if reactivated.Error != nil {
				return reactivated.Error
			}
			if reactivated.RowsAffected == 0 {
				return nil
			}
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

// LikeCommentAndMeta 点赞评论并原子自增评论点赞计数（事务，幂等，并发安全）。
// commentAuthor 为评论作者 ID：interacts.user_to 有外键约束，必须指向真实用户（不能为 0）。
// 并发安全依赖 interacts.single_key 唯一索引（同 LikeMomentAndMeta）：重复点赞被数据库原子忽略，
// 计数仅在真正新增或恢复为正常时 +1。
// 计数使用 ON DUPLICATE KEY UPDATE 自愈式自增：comment_interact_count 行在评论创建时不一定存在
// （功能上线前已存在的评论/楼中楼回复），普通 UpdateColumn 在行缺失时会静默丢失计数。
func (d *InteractsDAO) LikeCommentAndMeta(userID uint32, commentID uint64, commentAuthor uint32) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		interact := modelinteracts.Interacts{
			UserFrom:   userID,
			UserTo:     commentAuthor,
			TargetID:   commentID,
			Type:       consts.InteractTypeLike,
			TargetType: consts.InteractTargetComment,
			Status:     consts.InteractStatusNormal,
		}
		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&interact)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected == 0 {
			// 已有互动行：仅当之前不是正常状态时恢复为正常
			reactivated := tx.Model(&modelinteracts.Interacts{}).
				Where("user_from = ? AND target_id = ? AND type = ? AND target_type = ? AND status <> ?",
					userID, commentID, consts.InteractTypeLike, consts.InteractTargetComment, consts.InteractStatusNormal).
				Updates(map[string]interface{}{
					"status":   consts.InteractStatusNormal,
					"acted_at": time.Now(),
				})
			if reactivated.Error != nil {
				return reactivated.Error
			}
			if reactivated.RowsAffected == 0 {
				return nil
			}
		}
		return tx.Model(&modelinteracts.CommentInteractCount{}).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "comment_id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"like_count": gorm.Expr("like_count + ?", 1),
				}),
			}).
			Create(&modelinteracts.CommentInteractCount{CommentID: commentID, LikeCount: 1}).Error
	})
}

// UnlikeCommentAndMeta 取消评论点赞并原子自减评论点赞计数（事务，幂等）。
// 计数使用 GREATEST 下限保护，避免计数漂移时减成负数。
func (d *InteractsDAO) UnlikeCommentAndMeta(userID uint32, commentID uint64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&modelinteracts.Interacts{}).
			Where("user_from = ? AND target_id = ? AND type = ? AND target_type = ? AND status = ?",
				userID, commentID, consts.InteractTypeLike, consts.InteractTargetComment, consts.InteractStatusNormal).
			Update("status", consts.InteractStatusRevoked)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&modelinteracts.CommentInteractCount{}).
			Where("comment_id = ?", commentID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - ?, 0)", 1)).Error
	})
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

// HasLikedMoment 查询 userID 是否已点赞该动态（type=like, target_type=moment, status=normal）
func (d *InteractsDAO) HasLikedMoment(userID uint32, momentID uint64) (bool, error) {
	var count int64
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_from = ? AND target_id = ? AND type = ? AND target_type = ? AND status = ?",
			userID, momentID, consts.InteractTypeLike, consts.InteractTargetMoment, consts.InteractStatusNormal).
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

// QueryFollowingIDs 查询 userID 正在关注的全部用户 ID（user_from = userID，status 正常）。
// 用于关注流 feed 的作者集合。
func (d *InteractsDAO) QueryFollowingIDs(userID uint32) ([]uint32, error) {
	var ids []uint32
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_from = ? AND type = ? AND status = ?",
			userID, consts.InteractTypeFollow, consts.InteractStatusNormal).
		Pluck("user_to", &ids).Error
	return ids, err
}

// QueryFollowedByIDs 查询关注了 userID 的用户 ID 集合（user_to = userID，status 正常）。
// 用于关注流中"仅好友/仅粉丝"动态的可见性判定（作者是否回关查看者）。
func (d *InteractsDAO) QueryFollowedByIDs(userID uint32) ([]uint32, error) {
	var ids []uint32
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Where("user_to = ? AND type = ? AND status = ?",
			userID, consts.InteractTypeFollow, consts.InteractStatusNormal).
		Pluck("user_from", &ids).Error
	return ids, err
}

// QueryFollowStatusBatch 批量查询 viewerID 与各 targetID 的关注关系。
// 返回两个集合：viewerFollows（viewerID 关注了 targetID）、targetFollowsViewer（targetID 关注了 viewerID）。
// 一次 GROUP BY 查询，避免逐条 COUNT。
func (d *InteractsDAO) QueryFollowStatusBatch(viewerID uint32, targetIDs []uint32) (viewerFollows map[uint32]bool, targetFollowsViewer map[uint32]bool, err error) {
	viewerFollows = map[uint32]bool{}
	targetFollowsViewer = map[uint32]bool{}
	if len(targetIDs) == 0 {
		return viewerFollows, targetFollowsViewer, nil
	}

	var rows []struct {
		UserFrom uint32
		UserTo   uint32
	}
	err = d.DB.Model(&modelinteracts.Interacts{}).
		Select("DISTINCT user_from, user_to").
		Where("type = ? AND status = ? AND ((user_from = ? AND user_to IN ?) OR (user_to = ? AND user_from IN ?))",
			consts.InteractTypeFollow, consts.InteractStatusNormal,
			viewerID, targetIDs, viewerID, targetIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, nil, err
	}
	for _, row := range rows {
		if row.UserFrom == viewerID {
			viewerFollows[row.UserTo] = true
		}
		if row.UserTo == viewerID {
			targetFollowsViewer[row.UserFrom] = true
		}
	}
	return viewerFollows, targetFollowsViewer, nil
}

// HasLikedMomentsBatch 批量查询 userID 是否已点赞各动态（type=like, target_type=moment, status=normal）。
// 一次 GROUP BY 查询，避免逐条 COUNT。
func (d *InteractsDAO) HasLikedMomentsBatch(userID uint32, momentIDs []uint64) (map[uint64]bool, error) {
	result := map[uint64]bool{}
	if len(momentIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		TargetID uint64
	}
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Select("DISTINCT target_id").
		Where("user_from = ? AND target_id IN ? AND type = ? AND target_type = ? AND status = ?",
			userID, momentIDs, consts.InteractTypeLike, consts.InteractTargetMoment, consts.InteractStatusNormal).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.TargetID] = true
	}
	return result, nil
}

// HasLikedArticlesBatch 批量查询 userID 是否已点赞各文章（type=like, status=normal）。
// 注意：与现有文章详情接口（HasLikedMoment）口径一致，文章点赞记录复用 target_type=moment，
// 不引入新的 target_type，避免与既有数据不一致。
func (d *InteractsDAO) HasLikedArticlesBatch(userID uint32, articleIDs []uint64) (map[uint64]bool, error) {
	result := map[uint64]bool{}
	if len(articleIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		TargetID uint64
	}
	err := d.DB.Model(&modelinteracts.Interacts{}).
		Select("DISTINCT target_id").
		Where("user_from = ? AND target_id IN ? AND type = ? AND target_type = ? AND status = ?",
			userID, articleIDs, consts.InteractTypeLike, consts.InteractTargetMoment, consts.InteractStatusNormal).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.TargetID] = true
	}
	return result, nil
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
