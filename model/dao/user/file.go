package user

import "time"

const (
	FileStatusActive     uint8 = 1
	FileStatusProcessing uint8 = 2
	FileStatusFailed     uint8 = 3
	FileStatusDeleted    uint8 = 4
)

// FileType 文件大类分类（固定值，100% 写死不变，可用 uint8 常量）
const (
	FileTypeImage    uint8 = 1
	FileTypeVideo    uint8 = 2
	FileTypeAudio    uint8 = 3
	FileTypeDocument uint8 = 4
	FileTypeOther    uint8 = 5
)

// FilePermission 文件分享权限（与动态可见权限一致）
const (
	FilePermissionPublic  uint8 = 0 // 给全部人公开
	FilePermissionFriends uint8 = 1 // 给好友公开
	FilePermissionFans    uint8 = 3 // 给粉丝公开
	FilePermissionNone    uint8 = 2 // 不给任何人公开
)

type File struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	UUID         string    `gorm:"column:uuid;not null;uniqueIndex" json:"uuid"`
	UserID       uint32    `gorm:"column:user_id;not null;index" json:"user_id"`
	FileName     string    `gorm:"column:file_name;not null" json:"file_name"`
	ObjectKey    string    `gorm:"column:object_key;not null" json:"object_key"`
	FileURL      string    `gorm:"column:file_url;not null" json:"file_url"`
	FileType     uint8     `gorm:"column:file_type;not null" json:"file_type"`
	FileExt      string    `gorm:"column:file_ext;not null;default:''" json:"file_ext"`
	MimeType     string    `gorm:"column:mime_type;not null;default:''" json:"mime_type"`
	FileSize     uint64    `gorm:"column:file_size;not null;default:0" json:"file_size"`
	Permission   uint8     `gorm:"column:permission;not null;default:2" json:"permission"`
	Width        *uint32   `gorm:"column:width" json:"width,omitempty"`
	Height       *uint32   `gorm:"column:height" json:"height,omitempty"`
	Duration     *uint32   `gorm:"column:duration" json:"duration,omitempty"`
	ThumbnailURL *string   `gorm:"column:thumbnail_url" json:"thumbnail_url,omitempty"`
	Hash         [32]byte  `gorm:"column:hash;type:binary(32);not null;index" json:"-"`
	Status       uint8     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (File) TableName() string {
	return "files"
}
