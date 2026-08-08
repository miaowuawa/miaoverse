package user

import "time"

type Notify struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	UserID    uint32     `gorm:"not null;index" json:"user_id"`
	Type      uint8      `gorm:"not null" json:"type"`
	Content   string     `gorm:"not null;default:''" json:"content"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	ReadAt    *time.Time `gorm:"default:null" json:"read_at"`
	Received  bool       `gorm:"not null;default:false" json:"received"`
	Status    uint8      `gorm:"not null;default:0" json:"status"`
}

func (Notify) TableName() string {
	return "notify"
}
