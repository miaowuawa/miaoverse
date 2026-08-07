package moment

import "time"

const (
	MomentStatusNormal     uint8 = 0 // 正常
	MomentStatusDeleted    uint8 = 1 // 删除
	MomentStatusDraft      uint8 = 2 // 草稿
	MomentStatusRestricted uint8 = 3 // 限制传播
	MomentStatusBlocked    uint8 = 4 // 屏蔽
)

const (
	MomentPermissionPublic  uint8 = 0 // 公开
	MomentPermissionFriends uint8 = 1 // 仅好友
	MomentPermissionPrivate uint8 = 2 // 仅自己
	MomentPermissionFans    uint8 = 3 // 仅粉丝
)

const (
	MomentCommentPermissionAll     uint8 = 0 // 全部可以评论
	MomentCommentPermissionFriends uint8 = 1 // 仅好友（互相关注）可以评论
	MomentCommentPermissionFans    uint8 = 2 // 仅粉丝可以评论
	MomentCommentPermissionNone    uint8 = 3 // 全部不能评论
)

const (
	MomentTopNone     uint8 = 0   // 不置顶
	MomentTopPersonal uint8 = 1   // 个人-首页置顶
	MomentTopGlobal   uint8 = 100 // 全站-首页置顶
)

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

// MomentMetaData 动态计数元数据（冗余计数，由互动事件维护）
type MomentMetaData struct {
	MomentID     uint64    `gorm:"primaryKey" json:"moment_id"`
	LikeCount    uint32    `gorm:"not null;default:0" json:"like_count"`
	CommentCount uint32    `gorm:"not null;default:0" json:"comment_count"`
	ShareCount   uint32    `gorm:"not null;default:0" json:"share_count"`
	ViewCount    uint32    `gorm:"not null;default:0" json:"view_count"`
	ClickCount   uint32    `gorm:"not null;default:0" json:"click_count"`
	RepostCount  uint32    `gorm:"not null;default:0" json:"repost_count"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (MomentMetaData) TableName() string {
	return "moment_meta"
}
