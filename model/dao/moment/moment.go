package moment

import "time"

type Moment struct {
	ID                uint64     `gorm:"primaryKey" json:"id"`
	UserID            uint32     `gorm:"not null;index" json:"user_id"`
	Title             string     `gorm:"not null;default:''" json:"title"`
	Content           string     `gorm:"not null;default:''" json:"content"`
	Status            uint8      `gorm:"not null;default:0" json:"status"`
	Permission        uint8      `gorm:"not null;default:0" json:"permission"`
	CommentPermission uint8      `gorm:"not null;default:0" json:"comment_permission"`
	Top               uint8      `gorm:"not null;default:0" json:"top"`
	CreatedAt         time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt         *time.Time `gorm:"default:null" json:"deleted_at"`
}

func (Moment) TableName() string {
	return "moment"
}

// MomentInteractCount 动态互动计数（冗余计数，由互动事件维护）
type MomentInteractCount struct {
	MomentID     uint64    `gorm:"primaryKey" json:"moment_id"`
	LikeCount    uint32    `gorm:"not null;default:0" json:"like_count"`
	CommentCount uint32    `gorm:"not null;default:0" json:"comment_count"`
	ShareCount   uint32    `gorm:"not null;default:0" json:"share_count"`
	ViewCount    uint32    `gorm:"not null;default:0" json:"view_count"`
	ClickCount   uint32    `gorm:"not null;default:0" json:"click_count"`
	RepostCount  uint32    `gorm:"not null;default:0" json:"repost_count"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (MomentInteractCount) TableName() string {
	return "moment_interact_count"
}
