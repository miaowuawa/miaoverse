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
}

func (Interacts) TableName() string {
	return "interacts"
}
