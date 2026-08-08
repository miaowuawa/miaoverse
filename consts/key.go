package consts

// 中间件/会话/拉黑存储使用的 key 字符串。

const (
	UserLocalKey = "miaoverse.user"
	UIDLocalKey  = "miaoverse.uid"
)

const (
	BlockTargetLocalKey  = "miaoverse.block.target"
	BlockMomentLocalKey  = "miaoverse.block.moment"
	BlockCommentLocalKey = "miaoverse.block.comment"
	BlockCommentRootKey  = "miaoverse.block.comment.root"
)

const (
	SessionPhone              = "Phone"
	SessionRegion             = "Region"
	SessionUID                = "UID"
	SessionPendingLoginPhone  = "PendingLoginPhone"
	SessionPendingLoginRegion = "PendingLoginRegion"
)

const (
	BlockKeyPrefix = "block:user:"
	BlockKeySuffix = ":"
)
