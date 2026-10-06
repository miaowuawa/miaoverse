package sticker

import (
	"gorm.io/gorm"
	"miaoverse/consts"
	modelsticker "miaoverse/model/dao/sticker"
)

// StickerDAO 贴纸域DAO结构体，持有DB连接，挂载贴纸/贴纸包/收藏相关方法
type StickerDAO struct {
	DB *gorm.DB // 关联根DB，复用连接池
}

// ===== 贴纸 =====

// CreateSticker 创建贴纸记录。
func (d *StickerDAO) CreateSticker(s modelsticker.Sticker) (*modelsticker.Sticker, error) {
	if err := d.DB.Create(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// QueryStickerByUUID 按公开 UUID 查询 active 贴纸。
func (d *StickerDAO) QueryStickerByUUID(fileUUID string) (*modelsticker.Sticker, error) {
	var s modelsticker.Sticker
	err := d.DB.Where("uuid = ? AND status = ?", fileUUID, consts.StickerStatusActive).First(&s).Error
	return &s, err
}

// QueryStickersByUUIDsBatch 按公开 UUID 列表批量查询 active 贴纸。
// 返回 uuid → 贴纸记录；不存在的 UUID 不会出现在结果中。
func (d *StickerDAO) QueryStickersByUUIDsBatch(uuids []string) (map[string]modelsticker.Sticker, error) {
	result := map[string]modelsticker.Sticker{}
	if len(uuids) == 0 {
		return result, nil
	}
	var list []modelsticker.Sticker
	if err := d.DB.Where("uuid IN ? AND status = ?", uuids, consts.StickerStatusActive).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, s := range list {
		result[s.UUID] = s
	}
	return result, nil
}

// QueryStickersByIDsBatch 按 ID 列表批量查询 active 贴纸。
func (d *StickerDAO) QueryStickersByIDsBatch(ids []uint64) (map[uint64]modelsticker.Sticker, error) {
	result := map[uint64]modelsticker.Sticker{}
	if len(ids) == 0 {
		return result, nil
	}
	var list []modelsticker.Sticker
	if err := d.DB.Where("id IN ? AND status = ?", ids, consts.StickerStatusActive).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, s := range list {
		result[s.ID] = s
	}
	return result, nil
}

// QueryStickersByPack 查询贴纸包内全部 active 贴纸（按 sort 升序，同 sort 按 id 升序）。
func (d *StickerDAO) QueryStickersByPack(packID uint64) ([]modelsticker.Sticker, error) {
	var list []modelsticker.Sticker
	err := d.DB.Where("pack_id = ? AND status = ?", packID, consts.StickerStatusActive).
		Order("sort ASC, id ASC").
		Find(&list).Error
	return list, err
}

// CountStickersByPack 统计贴纸包内 active 贴纸数量。
func (d *StickerDAO) CountStickersByPack(packID uint64) (int64, error) {
	var count int64
	err := d.DB.Model(&modelsticker.Sticker{}).
		Where("pack_id = ? AND status = ?", packID, consts.StickerStatusActive).
		Count(&count).Error
	return count, err
}

// UpdateStickerPack 设置贴纸所属贴纸包（packID=0 表示移出贴纸包）。
func (d *StickerDAO) UpdateStickerPack(stickerID uint64, packID uint64, sort uint32) error {
	return d.DB.Model(&modelsticker.Sticker{}).Where("id = ?", stickerID).
		Updates(map[string]any{"pack_id": packID, "sort": sort}).Error
}

// DeleteSticker 软删除贴纸（status 置为 deleted）。
func (d *StickerDAO) DeleteSticker(stickerID uint64) error {
	return d.DB.Model(&modelsticker.Sticker{}).Where("id = ?", stickerID).
		Update("status", consts.StickerStatusDeleted).Error
}

// ===== 贴纸包 =====

// CreatePack 创建贴纸包。
func (d *StickerDAO) CreatePack(p modelsticker.Pack) (*modelsticker.Pack, error) {
	if err := d.DB.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// QueryPackByID 按 ID 查询 active 贴纸包。
func (d *StickerDAO) QueryPackByID(id uint64) (*modelsticker.Pack, error) {
	var p modelsticker.Pack
	err := d.DB.Where("id = ? AND status = ?", id, consts.StickerPackStatusActive).First(&p).Error
	return &p, err
}

// QueryPacksByIDsBatch 按 ID 列表批量查询 active 贴纸包（含封禁标记，供贴纸展示判定）。
func (d *StickerDAO) QueryPacksByIDsBatch(ids []uint64) (map[uint64]modelsticker.Pack, error) {
	result := map[uint64]modelsticker.Pack{}
	if len(ids) == 0 {
		return result, nil
	}
	var list []modelsticker.Pack
	if err := d.DB.Where("id IN ? AND status = ?", ids, consts.StickerPackStatusActive).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, p := range list {
		result[p.ID] = p
	}
	return result, nil
}

// QueryPacks 分页查询未封禁的贴纸包（按创建时间倒序）。
func (d *StickerDAO) QueryPacks(offset, limit int) ([]modelsticker.Pack, error) {
	var list []modelsticker.Pack
	err := d.DB.Where("status = ? AND banned = ?", consts.StickerPackStatusActive, consts.StickerPackBannedNone).
		Order("id DESC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// CountPacks 统计未封禁的贴纸包总数。
func (d *StickerDAO) CountPacks() (int64, error) {
	var count int64
	err := d.DB.Model(&modelsticker.Pack{}).
		Where("status = ? AND banned = ?", consts.StickerPackStatusActive, consts.StickerPackBannedNone).
		Count(&count).Error
	return count, err
}

// UpdatePackBanned 设置/解除贴纸包封禁标记（运营处置，与内容屏蔽状态一致由管理侧写入）。
func (d *StickerDAO) UpdatePackBanned(packID uint64, banned uint8) error {
	return d.DB.Model(&modelsticker.Pack{}).Where("id = ?", packID).
		Update("banned", banned).Error
}

// ===== 用户贴纸收藏夹 =====

// CreateUserSticker 创建收藏夹条目。
func (d *StickerDAO) CreateUserSticker(us modelsticker.UserSticker) (*modelsticker.UserSticker, error) {
	if err := d.DB.Create(&us).Error; err != nil {
		return nil, err
	}
	return &us, nil
}

// QueryUserSticker 查询收藏夹条目（不存在返回 gorm.ErrRecordNotFound）。
func (d *StickerDAO) QueryUserSticker(userID uint32, stickerID uint64) (*modelsticker.UserSticker, error) {
	var us modelsticker.UserSticker
	err := d.DB.Where("user_id = ? AND sticker_id = ?", userID, stickerID).First(&us).Error
	return &us, err
}

// QueryUserStickerIDsBatch 批量查询用户收藏夹中已存在的贴纸 ID → 条目。
func (d *StickerDAO) QueryUserStickerIDsBatch(userID uint32, stickerIDs []uint64) (map[uint64]modelsticker.UserSticker, error) {
	result := map[uint64]modelsticker.UserSticker{}
	if len(stickerIDs) == 0 {
		return result, nil
	}
	var list []modelsticker.UserSticker
	if err := d.DB.Where("user_id = ? AND sticker_id IN ?", userID, stickerIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, us := range list {
		result[us.StickerID] = us
	}
	return result, nil
}

// DeleteUserSticker 删除收藏夹条目（本人上传的贴纸条目随贴纸删除一并清理）。
func (d *StickerDAO) DeleteUserSticker(userID uint32, stickerID uint64) error {
	return d.DB.Where("user_id = ? AND sticker_id = ?", userID, stickerID).
		Delete(&modelsticker.UserSticker{}).Error
}

// CountUserStickers 统计用户收藏夹条目数；source=0 统计全部来源。
func (d *StickerDAO) CountUserStickers(userID uint32, source uint8) (int64, error) {
	q := d.DB.Model(&modelsticker.UserSticker{}).Where("user_id = ?", userID)
	if source != 0 {
		q = q.Where("source = ?", source)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

// QueryUserStickers 分页查询用户收藏夹条目（置顶优先，其次按加入时间升序）。
func (d *StickerDAO) QueryUserStickers(userID uint32, offset, limit int) ([]modelsticker.UserSticker, error) {
	var list []modelsticker.UserSticker
	err := d.DB.Where("user_id = ?", userID).
		Order("top DESC, id ASC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// UpdateUserStickerTop 设置/取消收藏夹贴纸置顶。
func (d *StickerDAO) UpdateUserStickerTop(userID uint32, stickerID uint64, top uint8) error {
	return d.DB.Model(&modelsticker.UserSticker{}).Where("user_id = ? AND sticker_id = ?", userID, stickerID).
		Update("top", top).Error
}

// ===== 贴纸包收藏 =====

// CreatePackFavorite 收藏贴纸包（幂等由调用方先查后插）。
func (d *StickerDAO) CreatePackFavorite(userID uint32, packID uint64) error {
	return d.DB.Create(&modelsticker.PackFavorite{UserID: userID, PackID: packID}).Error
}

// DeletePackFavorite 取消收藏贴纸包（幂等）。
func (d *StickerDAO) DeletePackFavorite(userID uint32, packID uint64) error {
	return d.DB.Where("user_id = ? AND pack_id = ?", userID, packID).
		Delete(&modelsticker.PackFavorite{}).Error
}

// QueryPackFavorite 查询贴纸包收藏记录（不存在返回 gorm.ErrRecordNotFound）。
func (d *StickerDAO) QueryPackFavorite(userID uint32, packID uint64) (*modelsticker.PackFavorite, error) {
	var pf modelsticker.PackFavorite
	err := d.DB.Where("user_id = ? AND pack_id = ?", userID, packID).First(&pf).Error
	return &pf, err
}

// QueryFavoritePackIDs 查询用户已收藏的贴纸包 ID 集合。
func (d *StickerDAO) QueryFavoritePackIDs(userID uint32) (map[uint64]bool, error) {
	result := map[uint64]bool{}
	var rows []struct {
		PackID uint64
	}
	if err := d.DB.Model(&modelsticker.PackFavorite{}).
		Select("pack_id").
		Where("user_id = ?", userID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.PackID] = true
	}
	return result, nil
}

// QueryFavoritePacks 分页查询用户已收藏的贴纸包（按收藏时间倒序）。
func (d *StickerDAO) QueryFavoritePacks(userID uint32, offset, limit int) ([]modelsticker.Pack, error) {
	var list []modelsticker.Pack
	err := d.DB.Table("sticker_pack p").
		Select("p.*").
		Joins("INNER JOIN sticker_pack_favorite f ON f.pack_id = p.id AND f.user_id = ?", userID).
		Where("p.status = ?", consts.StickerPackStatusActive).
		Order("f.id DESC").
		Offset(offset).Limit(limit).
		Scan(&list).Error
	return list, err
}

// CountFavoritePacks 统计用户已收藏的贴纸包数量。
func (d *StickerDAO) CountFavoritePacks(userID uint32) (int64, error) {
	var count int64
	err := d.DB.Model(&modelsticker.PackFavorite{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// QueryFavoritedPackStickerIDsBatch 批量查询贴纸是否属于用户已收藏贴纸包（可用于评论贴纸使用权限校验）。
// 返回 stickerID → true（属于任一已收藏贴纸包）。
func (d *StickerDAO) QueryFavoritedPackStickerIDsBatch(userID uint32, stickerIDs []uint64) (map[uint64]bool, error) {
	result := map[uint64]bool{}
	if len(stickerIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		StickerID uint64
	}
	err := d.DB.Table("sticker s").
		Select("DISTINCT s.id AS sticker_id").
		Joins("INNER JOIN sticker_pack_favorite f ON f.pack_id = s.pack_id AND f.user_id = ?", userID).
		Where("s.id IN ?", stickerIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.StickerID] = true
	}
	return result, nil
}
