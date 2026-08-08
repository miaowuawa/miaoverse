package consts

// 动态（moment）域常量：状态、可见权限、评论权限、置顶状态、长度限制、内容类型标识。

const (
	MomentStatusNormal     uint8 = 0 // 正常
	MomentStatusDeleted    uint8 = 1 // 删除
	MomentStatusDraft      uint8 = 2 // 草稿
	MomentStatusRestricted uint8 = 3 // 限制传播
	MomentStatusBlocked    uint8 = 4 // 屏蔽
)

const (
	MomentPermissionPublic  uint8 = 0 // 公开
	MomentPermissionFriends uint8 = 1 // 仅好友
	MomentPermissionPrivate uint8 = 2 // 仅自己
	MomentPermissionFans    uint8 = 3 // 仅粉丝
)

const (
	MomentCommentPermissionAll     uint8 = 0 // 全部可以评论
	MomentCommentPermissionFriends uint8 = 1 // 仅好友（互相关注）可以评论
	MomentCommentPermissionFans    uint8 = 2 // 仅粉丝可以评论
	MomentCommentPermissionNone    uint8 = 3 // 全部不能评论
)

const (
	MomentTopNone     uint8 = 0   // 不置顶
	MomentTopPersonal uint8 = 1   // 个人-首页置顶
	MomentTopGlobal   uint8 = 100 // 全站-首页置顶
)

const (
	MaxMomentContentLen = 5000
	MaxMomentTitleLen   = 255

	ContentTypeMoment = "moment" // 内容类型标识
)
