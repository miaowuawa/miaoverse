package interacts

import "time"

const (
	CommentTargetMoment  uint8 = 1 // 动态
	CommentTargetComment uint8 = 2 // 其他评论（楼中楼）
)

const (
	CommentStatusNormal     uint8 = 0 // 正常
	CommentStatusDeleted    uint8 = 1 // 删除
	CommentStatusDraft      uint8 = 2 // 草稿
	CommentStatusRestricted uint8 = 3 // 限制传播
	CommentStatusBlocked    uint8 = 4 // 屏蔽
)

type Comment struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	UserID     uint32     `gorm:"not null;index" json:"user_id"`
	TargetID   uint64     `gorm:"not null;index" json:"target_id"`
	TargetType uint8      `gorm:"not null" json:"target_type"`
	Content    string     `gorm:"not null;default:''" json:"content"`
	Status     uint8      `gorm:"not null;default:0" json:"status"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"default:null" json:"deleted_at"`
}

func (Comment) TableName() string {
	return "comment"
}
