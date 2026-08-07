package user

import "time"

// 权限位定义（bitmask）。1 表示该权限被封禁，0 表示未封禁。
const (
	PermComment        = 1 << 0 // 1    发表评论
	PermPost           = 1 << 1 // 2    发布动态
	PermPrivate        = 1 << 2 // 4    主动私信
	PermAvatar         = 1 << 3 // 8    更改头像
	PermNickname       = 1 << 4 // 16   更改昵称
	PermSignature      = 1 << 5 // 32   更改个性签名
	PermSocial         = 1 << 6 // 64   社交互动（转发/点赞/关注）
	PermDeleteRegister = 1 << 7 // 128  注销和注册账号
	PermUploadFile     = 1 << 8 // 256  上传文件
)

const (
	PunishmentStatusActive  uint8 = 1 // 生效中
	PunishmentStatusEnded   uint8 = 2 // 已到期
	PunishmentStatusRevoked uint8 = 3 // 已撤销
)

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
