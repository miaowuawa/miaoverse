package consts

// 用户域常量：账号状态、权限位、惩罚状态、资料长度限制。

const (
	UserStatusActive   uint8 = 1 // 正常
	UserStatusBanned   uint8 = 2 // 封禁
	UserStatusClosed   uint8 = 3 // 注销
	UserStatusDisabled uint8 = 4 // 停用
)

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

const (
	MaxUsernameLen = 64
	MaxNicknameLen = 64
	MaxAvatarLen   = 255
	MaxBioLen      = 255
)
