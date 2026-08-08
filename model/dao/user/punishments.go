package user

import "time"

type Punishment struct {
	ID                 uint64     `gorm:"primaryKey" json:"id"`
	UserID             uint32     `gorm:"column:user_id;not null;index" json:"user_id"`
	PunishmentSwitch   uint32     `gorm:"column:punishment_type;not null;default:0" json:"punishment_type"`
	PunishmentStatus   uint8      `gorm:"column:punishment_status;not null;default:1" json:"punishment_status"`
	PunishmentTime     time.Time  `gorm:"column:punishment_time;not null;default:CURRENT_TIMESTAMP" json:"punishment_time"`
	PunishmentEndTime  *time.Time `gorm:"column:punishment_end_time" json:"punishment_end_time"`
	PunishmentReason   string     `gorm:"column:punishment_reason;not null;default:''" json:"punishment_reason"`
	PunishmentOperator uint32     `gorm:"column:punishment_operator;not null;default:0" json:"punishment_operator"`
	PunishmentRemark   string     `gorm:"column:punishment_remark;not null;default:''" json:"punishment_remark"`
}

func (Punishment) TableName() string {
	return "punishment"
}
