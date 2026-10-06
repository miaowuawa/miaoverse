package stickerreq

// CreateStickerPack 创建贴纸包请求体。
type CreateStickerPack struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddStickerToPack 向贴纸包添加贴纸请求体（贴纸必须属于当前用户）。
type AddStickerToPack struct {
	StickerUUID string `json:"sticker_uuid"`
}

// FavoriteStickerPack 收藏/取消收藏贴纸包请求体。
type FavoriteStickerPack struct {
	PackID uint64 `json:"pack_id"`
}
