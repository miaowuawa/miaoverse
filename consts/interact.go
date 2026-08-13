package consts

// 互动域常量：互动类型、目标类型、状态、动作字符串、分页限制。

const (
	InteractTypeFollow   uint8 = 0   // 关注
	InteractTypeLike     uint8 = 1   // 点赞
	InteractTypeShare    uint8 = 2   // 分享
	InteractTypeRepost   uint8 = 3   // 转发
	InteractTypeFavorite uint8 = 4   // 收藏
	InteractTypeDMFirst  uint8 = 100 // 首次主动私信
	InteractTypeDMReply  uint8 = 101 // 回复主动私信
	InteractTypeDMNormal uint8 = 102 // 回复普通私信
	InteractTypeComment  uint8 = 103 // 对内容评论
	InteractTypeReply    uint8 = 104 // 回复评论
)

const (
	InteractTargetUser    uint8 = 0 // 用户
	InteractTargetMoment  uint8 = 1 // 动态
	InteractTargetComment uint8 = 2 // 评论
	InteractTargetReply   uint8 = 3 // 回复
)

const (
	InteractStatusNormal  uint8 = 0  // 正常
	InteractStatusRevoked uint8 = 9  // 已撤销（自主）
	InteractStatusForced  uint8 = 10 // 已撤销（非自愿）
)

const (
	ActionAdd      = "add"
	ActionRemove   = "remove"
	ActionFollow   = "follow"
	ActionUnfollow = "unfollow"
	ActionLike     = "like"
	ActionUnlike   = "unlike"
)

const (
	DefaultListLimit = 20
	MaxListLimit     = 100
)
