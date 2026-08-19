package interacts

import "time"

type Interacts struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	UserFrom    uint32    `gorm:"not null;index" json:"user_from"`
	UserTo      uint32    `gorm:"not null;index" json:"user_to"`
	TargetID    uint64    `gorm:"not null" json:"target_id"`
	ReferenceID uint64    `gorm:"not null;default:0" json:"reference_id"`
	Type        uint8     `gorm:"not null" json:"type"`
	TargetType  uint8     `gorm:"not null" json:"target_type"`
	Status      uint8     `gorm:"not null;default:0" json:"status"`
	ActedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"acted_at"`
	// SingleKey 数据库生成列（只读，创建时不写入）：单实例互动（关注/点赞/分享/转发/收藏）的唯一键，
	// 由 uk_interacts_single_key 唯一索引保证同一用户对同一目标最多一行，从数据库层面消除并发重复。
	// 评论/回复/私信等可多次操作的类型该列为 NULL（NULL 不参与唯一约束）。
	SingleKey string `gorm:"->;column:single_key" json:"-"`
}

func (Interacts) TableName() string {
	return "interacts"
}
