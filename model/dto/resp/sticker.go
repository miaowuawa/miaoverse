package resp

// StickerInfo 贴纸信息。FileUUID 为贴纸图片文件 UUID（原始存储 URL 不下发，需经临时链接接口换取）。
// Source/Top 为收藏夹视角字段：仅贴纸收藏夹列表返回（1 本人上传，2 收藏；Top 1 为置顶），其余场景为 0。
type StickerInfo struct {
	UUID      string `json:"uuid"`
	FileUUID  string `json:"file_uuid"`
	PackID    uint64 `json:"pack_id"`
	Name      string `json:"name"`
	Source    uint8  `json:"source"`
	Top       uint8  `json:"top"`
	CreatedAt string `json:"created_at"`
}

type CodeWithMsgSticker struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Sticker StickerInfo `json:"sticker"`
}

type CodeWithMsgStickerList struct {
	Code     int           `json:"code"`
	Msg      string        `json:"msg"`
	Count    int64         `json:"count"`
	Stickers []StickerInfo `json:"stickers"`
}

// StickerPackInfo 贴纸包信息。StickerCount 为包内贴纸数，IsFavorite 为当前用户是否已收藏整包。
type StickerPackInfo struct {
	ID           uint64 `json:"id"`
	UUID         string `json:"uuid"`
	UserID       uint32 `json:"user_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Banned       bool   `json:"banned"`
	StickerCount int64  `json:"sticker_count"`
	IsFavorite   bool   `json:"is_favorite"`
	CreatedAt    string `json:"created_at"`
}

type CodeWithMsgStickerPack struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Pack StickerPackInfo `json:"pack"`
}

type CodeWithMsgStickerPackList struct {
	Code  int               `json:"code"`
	Msg   string            `json:"msg"`
	Count int64             `json:"count"`
	Packs []StickerPackInfo `json:"packs"`
}

// CodeWithMsgStickerPackDetail 贴纸包详情：包信息 + 包内贴纸列表。
type CodeWithMsgStickerPackDetail struct {
	Code     int             `json:"code"`
	Msg      string          `json:"msg"`
	Pack     StickerPackInfo `json:"pack"`
	Stickers []StickerInfo   `json:"stickers"`
}

// CommentSticker 评论内嵌贴纸的展示信息。
// Hidden=true 表示贴纸不可显示（所在贴纸包被封禁或贴纸已删除），
// 前端隐藏贴纸并在该评论下以灰字提示「部分贴纸未显示」。
type CommentSticker struct {
	UUID     string `json:"uuid"`
	FileUUID string `json:"file_uuid"`
	Name     string `json:"name"`
	Hidden   bool   `json:"hidden"`
}
