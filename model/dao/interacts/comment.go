package interacts

import "time"

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
