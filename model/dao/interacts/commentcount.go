package interacts

import "time"

// CommentInteractCount 评论互动计数（冗余计数，由互动事件维护）
type CommentInteractCount struct {
	CommentID uint64    `gorm:"primaryKey" json:"comment_id"`
	LikeCount uint32    `gorm:"not null;default:0" json:"like_count"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CommentInteractCount) TableName() string {
	return "comment_interact_count"
}
