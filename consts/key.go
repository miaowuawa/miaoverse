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
	BlockArticleMetaKey  = "miaoverse.block.article.meta"
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

// 通知/在线状态的 Redis key 与频道（跨实例共享，支撑无状态化部署）。
const (
	// NotifyPresenceKey 用户在线状态 ZSET key：member=用户 id，score=最近心跳/事件成功写出的 Unix 毫秒。
	NotifyPresenceKey = "notify:presence"
	// NotifyBusChannel 通知事件总线 Redis pub/sub 频道：通知事件跨实例广播，各实例投递到本实例的 SSE 连接。
	NotifyBusChannel = "notify:bus"
)
