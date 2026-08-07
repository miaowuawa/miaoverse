package user

import "time"

const (
	NotifyTypeAccountSecurity uint8 = 0 // 账号安全
	NotifyTypeTransaction     uint8 = 1 // 事务事项
	NotifyTypeLike            uint8 = 2 // 互动-赞
	NotifyTypeFollow          uint8 = 3 // 互动-关注
	NotifyTypeMention         uint8 = 4 // 互动-@我
	NotifyTypeReply           uint8 = 5 // 互动-回复与评论
)

const (
	NotifyStatusUnread  uint8 = 0 // 未读
	NotifyStatusRead    uint8 = 1 // 已读
	NotifyStatusDeleted uint8 = 2 // 已删除
)

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
