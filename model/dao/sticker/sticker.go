package sticker

import "time"

// Sticker 贴纸（单张贴纸图片）。贴纸图片存储在 files 表（file_uuid），原始存储 URL 不下发。
// PackID 为 0 表示未加入任何贴纸包；一张贴纸至多属于一个贴纸包。
type Sticker struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UUID      string    `gorm:"column:uuid;not null;uniqueIndex" json:"uuid"`
	UserID    uint32    `gorm:"column:user_id;not null;index" json:"user_id"`
	PackID    uint64    `gorm:"column:pack_id;not null;default:0;index" json:"pack_id"`
	FileUUID  string    `gorm:"column:file_uuid;not null" json:"file_uuid"`
	Name      string    `gorm:"column:name;not null;default:''" json:"name"`
	Sort      uint32    `gorm:"column:sort;not null;default:0" json:"sort"`
	Status    uint8     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Sticker) TableName() string {
	return "sticker"
}

// Pack 贴纸包。Banned 为封禁标记：封禁后包内贴纸在所有使用处无法显示。
type Pack struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	UUID        string    `gorm:"column:uuid;not null;uniqueIndex" json:"uuid"`
	UserID      uint32    `gorm:"column:user_id;not null;index" json:"user_id"`
	Name        string    `gorm:"column:name;not null;default:''" json:"name"`
	Description string    `gorm:"column:description;not null;default:''" json:"description"`
	Banned      uint8     `gorm:"column:banned;not null;default:0" json:"banned"`
	Status      uint8     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Pack) TableName() string {
	return "sticker_pack"
}

// UserSticker 用户贴纸收藏夹条目：本人上传的贴纸（source=1）与收藏他人贴纸（source=2）统一存放，
// Top 为置顶标记（收藏夹内优先展示）。
type UserSticker struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint32    `gorm:"column:user_id;not null;index" json:"user_id"`
	StickerID uint64    `gorm:"column:sticker_id;not null" json:"sticker_id"`
	Source    uint8     `gorm:"column:source;not null;default:1" json:"source"`
	Top       uint8     `gorm:"column:top;not null;default:0" json:"top"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (UserSticker) TableName() string {
	return "user_sticker"
}

// PackFavorite 贴纸包收藏：收藏整个贴纸包后，包内容变化自动同步（包内容实时查询）。
type PackFavorite struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint32    `gorm:"column:user_id;not null" json:"user_id"`
	PackID    uint64    `gorm:"column:pack_id;not null" json:"pack_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (PackFavorite) TableName() string {
	return "sticker_pack_favorite"
}
