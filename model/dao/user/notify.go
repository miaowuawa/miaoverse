package user

import "time"

// Notify 用户通知记录（notify 表）。
// Type 取值见 consts.NotifyType*；TargetType/TargetID 定位通知关联的对象
// （复用 consts.InteractTarget*：0 用户、1 动态、2 评论），ActorID 为触发通知的用户
// （账号安全类通知无触发者，为 0）。Status 取值见 consts.NotifyStatus*。
type Notify struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	UserID     uint32     `gorm:"not null;index" json:"user_id"`
	Type       uint8      `gorm:"not null" json:"type"`
	ActorID    uint32     `gorm:"not null;default:0" json:"actor_id"`
	TargetType uint8      `gorm:"not null;default:0" json:"target_type"`
	TargetID   uint64     `gorm:"not null;default:0" json:"target_id"`
	Content    string     `gorm:"not null;default:''" json:"content"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	ReadAt     *time.Time `gorm:"default:null" json:"read_at"`
	Received   bool       `gorm:"not null;default:false" json:"received"`
	Status     uint8      `gorm:"not null;default:0" json:"status"`
}

func (Notify) TableName() string {
	return "notify"
}
