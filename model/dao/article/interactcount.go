package article

import "time"

// ArticleInteractCount 文章互动计数（MySQL，冗余计数，由互动事件维护）。
// 与动态一致：点赞/评论/分享/浏览/点击/转发。文章按章节存储时，每章一条独立计数记录。
type ArticleInteractCount struct {
	ArticleID    uint64    `gorm:"primaryKey" json:"article_id"`
	LikeCount    uint32    `gorm:"not null;default:0" json:"like_count"`
	CommentCount uint32    `gorm:"not null;default:0" json:"comment_count"`
	ShareCount   uint32    `gorm:"not null;default:0" json:"share_count"`
	ViewCount    uint32    `gorm:"not null;default:0" json:"view_count"`
	ClickCount   uint32    `gorm:"not null;default:0" json:"click_count"`
	RepostCount  uint32    `gorm:"not null;default:0" json:"repost_count"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (ArticleInteractCount) TableName() string {
	return "article_interact_count"
}
