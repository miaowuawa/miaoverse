package resp

import (
	"time"

	"github.com/gofiber/fiber/v3"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/service/i18n"
)

// CodeWithMsgNotifyList 通知列表响应体（分页，按分类过滤）。
type CodeWithMsgNotifyList struct {
	Code     int          `json:"code"`
	Msg      string       `json:"msg"`
	Count    int64        `json:"count"`
	Notifies []NotifyInfo `json:"notifies"`
}

// CodeWithMsgNotifyUnread 未读数响应体。
type CodeWithMsgNotifyUnread struct {
	Code   int          `json:"code"`
	Msg    string       `json:"msg"`
	Unread NotifyUnread `json:"unread"`
}

// CodeWithMsgNotifyRead 标记已读/全部已读/删除响应体。
// 返回最新未读数，供前端同步各分类小红点。
type CodeWithMsgNotifyRead struct {
	Code   int          `json:"code"`
	Msg    string       `json:"msg"`
	Unread NotifyUnread `json:"unread"`
}

// NotifyInfo 通知信息。
// Category 为前端分类（account/like/follow/mention/reply，见 consts.NotifyCategory*），
// Actor 为触发通知的用户（账号安全类通知无触发者，为 nil），
// TargetType/TargetID 为通知关联对象（复用 consts.InteractTarget*：0 用户、1 动态、2 评论），
// Read 表示是否已读。已注销的触发者展示字段统一打码。
type NotifyInfo struct {
	ID         uint64          `json:"id"`
	Type       uint8           `json:"type"`
	Category   string          `json:"category"`
	Actor      *modeluser.User `json:"actor,omitempty"`
	TargetType uint8           `json:"target_type"`
	TargetID   uint64          `json:"target_id"`
	Content    string          `json:"content"`
	CreatedAt  time.Time       `json:"created_at"`
	ReadAt     *time.Time      `json:"read_at"`
	Read       bool            `json:"read"`
}

// NotifyUnread 各分类未读数（Total 为全部未读）。
type NotifyUnread struct {
	Total   int64 `json:"total"`
	Account int64 `json:"account"`
	Like    int64 `json:"like"`
	Follow  int64 `json:"follow"`
	Mention int64 `json:"mention"`
	Reply   int64 `json:"reply"`
}

// NotifyEvent SSE 推送事件载荷（event: notification）。
// Unread 为该通知写入后的最新未读数，前端收到后可直接刷新小红点。
type NotifyEvent struct {
	Notify NotifyInfo   `json:"notify"`
	Unread NotifyUnread `json:"unread"`
}

// PresenceInfo 用户在线状态（基于 SSE 连接心跳判定）。
// Online 为 true 表示有活跃 SSE 连接且心跳正常；LastSeen 为最近活跃时间（从未连接过为零值）。
type PresenceInfo struct {
	UID      uint32    `json:"uid"`
	Online   bool      `json:"online"`
	LastSeen time.Time `json:"last_seen"`
}

// CodeWithMsgPresence 在线状态批量查询响应体。
type CodeWithMsgPresence struct {
	Code     int            `json:"code"`
	Msg      string         `json:"msg"`
	Presence []PresenceInfo `json:"presence"`
}

// NotifyList 通知列表响应，已注销触发者的展示字段统一打码。
func NotifyList(ctx fiber.Ctx, count int64, notifies []NotifyInfo) error {
	for i := range notifies {
		MaskClosedAccount(ctx, notifies[i].Actor)
	}
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgNotifyList{
		Code:     fiber.StatusOK,
		Msg:      i18n.Message(ctx, i18n.OKNotifyList),
		Count:    count,
		Notifies: notifies,
	})
}

// NotifyUnreadOK 未读数查询成功响应。
func NotifyUnreadOK(ctx fiber.Ctx, unread NotifyUnread) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgNotifyUnread{
		Code:   fiber.StatusOK,
		Msg:    i18n.Message(ctx, i18n.OKNotifyFetched),
		Unread: unread,
	})
}

// NotifyReadOK 标记已读/全部已读成功响应。
func NotifyReadOK(ctx fiber.Ctx, unread NotifyUnread) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgNotifyRead{
		Code:   fiber.StatusOK,
		Msg:    i18n.Message(ctx, i18n.OKNotifyRead),
		Unread: unread,
	})
}

// NotifyDeletedOK 删除通知成功响应。
func NotifyDeletedOK(ctx fiber.Ctx, unread NotifyUnread) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgNotifyRead{
		Code:   fiber.StatusOK,
		Msg:    i18n.Message(ctx, i18n.OKNotifyDeleted),
		Unread: unread,
	})
}

// PresenceOK 在线状态批量查询成功响应。
func PresenceOK(ctx fiber.Ctx, presence []PresenceInfo) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgPresence{
		Code:     fiber.StatusOK,
		Msg:      i18n.Message(ctx, i18n.OKPresenceFetched),
		Presence: presence,
	})
}
