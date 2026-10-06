package consts

import "time"

// 通知域常量：类型、状态、前端分类、内容限制、SSE 推送参数。

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

// 通知前端分类（API 的 category 参数与响应字段取值）。
// 一个分类可对应多个通知类型，例如「账号」分类同时覆盖账号安全与事务事项。
const (
	NotifyCategoryAccount = "account" // 账号（NotifyTypeAccountSecurity、NotifyTypeTransaction）
	NotifyCategoryLike    = "like"    // 点赞（NotifyTypeLike）
	NotifyCategoryFollow  = "follow"  // 关注（NotifyTypeFollow）
	NotifyCategoryMention = "mention" // 提及（NotifyTypeMention）
	NotifyCategoryReply   = "reply"   // 回复与评论（NotifyTypeReply）
)

// MaxNotifyContentLen 通知落库时 content 字段允许的最大长度（rune），超长摘要在写入前截断。
const MaxNotifyContentLen = 200

// SSE 推送参数：
// NotifySSEHeartbeat 心跳帧间隔，用于保活并及时感知客户端断开；
// NotifySSEBufferSize 每个 SSE 连接的事件缓冲长度，缓冲写满时丢弃该事件（数据仍以数据库为准，前端可拉取补齐）；
// NotifySSERetryMs 断线后 EventSource 自动重连间隔提示（毫秒）。
const (
	NotifySSEHeartbeat  = 15 * time.Second
	NotifySSEBufferSize = 16
	NotifySSERetryMs    = 5000
)

// 在线状态判定（基于 SSE 连接心跳）：
// NotifySSEOfflineAfter 离线判定窗口——距最近一次成功写入（心跳/事件）超过该时长视为离线。
// 心跳每 NotifySSEHeartbeat 写出一次，写出失败立即断开连接，因此正常连接的活跃间隔不超过一个心跳周期，
// 两倍心跳的窗口既能容忍单次延迟，又能保证离线判定滞后不超过 NotifySSEOfflineAfter；
// NotifyMaxPresenceQueryUids 在线状态批量查询接口单次最多查询的用户数；
// NotifyPresenceRetention 在线状态记录保留时长：超过该时长未活跃的记录被清理（不影响在线判定窗口）。
const (
	NotifySSEOfflineAfter      = 2 * NotifySSEHeartbeat
	NotifyMaxPresenceQueryUids = 100
	NotifyPresenceRetention    = 7 * 24 * time.Hour
)
