package stickerreq

// FavoriteSticker 添加收藏贴纸请求体。
type FavoriteSticker struct {
	StickerUUID string `json:"sticker_uuid"`
}

// SetStickerTop 设置收藏夹贴纸置顶请求体（top：0 取消置顶，1 置顶）。
type SetStickerTop struct {
	Top uint8 `json:"top"`
}
