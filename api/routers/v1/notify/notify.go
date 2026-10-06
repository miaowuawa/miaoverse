package notify

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/i18n"
	"miaoverse/util/pagination"
)

// ListHandler 分页查询当前用户的通知（需登录）。
// query 参数：category 分类（account/like/follow/mention/reply，缺省为全部）、offset、limit。
func ListHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	offset, limit, ok := pagination.Parse(ctx.Query("offset"), ctx.Query("limit"))
	if !ok {
		return resp.BadRequest(ctx)
	}

	items, count, ok, err := servants.NotifyServant.List(uid, ctx.Query("category"), offset, limit, i18n.LanguageFromCtx(ctx))
	if err != nil {
		return resp.ServerError(ctx)
	}
	if !ok {
		return resp.BadRequest(ctx)
	}
	return resp.NotifyList(ctx, count, items)
}

// UnreadCountHandler 查询当前用户各分类未读数（需登录），供导航小红点展示。
func UnreadCountHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	unread, err := servants.NotifyServant.Unread(uid)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.NotifyUnreadOK(ctx, unread)
}

// MarkReadHandler 标记单条通知已读（幂等，需登录），返回最新未读数。
func MarkReadHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	id, ok := parseNotifyID(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	unread, err := servants.NotifyServant.MarkRead(uid, id)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.NotifyReadOK(ctx, unread)
}

// MarkAllReadHandler 将当前用户全部未读通知标记为已读（幂等，需登录），返回最新未读数。
func MarkAllReadHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	unread, err := servants.NotifyServant.MarkAllRead(uid)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.NotifyReadOK(ctx, unread)
}

// DeleteHandler 删除单条通知（软删除，幂等，需登录），返回最新未读数。
func DeleteHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	id, ok := parseNotifyID(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	unread, err := servants.NotifyServant.Delete(uid, id)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.NotifyDeletedOK(ctx, unread)
}

// StreamHandler 通知 SSE 实时推送流（需登录）。
//
// 建立 SSE 长连接（text/event-stream）后：
//   - 有新通知时推送 `event: notification`，data 为 resp.NotifyEvent（通知 + 最新未读数）；
//   - 每 consts.NotifySSEHeartbeat 下发一帧 `event: ping` 心跳：客户端据此判定连接在线
//     （收不到心跳即判离线并重连），服务端写出成功即刷新该用户在线活跃时间
//     （在线状态存 Redis 跨实例共享，判定规则见 Notify.PresenceStore）；
//   - 连接建立时下发 retry 提示，前端 EventSource 断线后按该间隔自动重连。
//
// 登录态通过 session cookie 校验（EventSource 同源请求自动携带），无需额外 ticket。
func StreamHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	// SSE 响应头：禁用缓冲与缓存，反向代理（X-Accel-Buffering）同样直通
	ctx.Set(fiber.HeaderContentType, "text/event-stream; charset=utf-8")
	ctx.Set(fiber.HeaderCacheControl, "no-cache")
	ctx.Set(fiber.HeaderConnection, "keep-alive")
	ctx.Set("X-Accel-Buffering", "no")

	sub := servants.NotifyServant.Subscribe(uid)

	return ctx.SendStreamWriter(func(w *bufio.Writer) {
		defer sub.Close()

		// 断线重连间隔提示（毫秒）+ 首帧确认
		if _, err := fmt.Fprintf(w, "retry: %d\n\n: connected\n\n", consts.NotifySSERetryMs); err != nil {
			return
		}
		if err := w.Flush(); err != nil {
			return
		}

		heartbeat := time.NewTicker(consts.NotifySSEHeartbeat)
		defer heartbeat.Stop()

		for {
			select {
			case ev, ok := <-sub.Events:
				if !ok {
					// 服务停机或连接已退订
					return
				}
				data, err := json.Marshal(ev)
				if err != nil {
					log.Printf("Notify.Stream: 事件序列化失败 user=%d: %v", uid, err)
					continue
				}
				if _, err := fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
				// 事件写出成功：连接存活，刷新在线活跃时间（Redis，跨实例共享）
				servants.NotifyServant.Heartbeat(uid)
			case <-heartbeat.C:
				if _, err := fmt.Fprintf(w, "event: ping\ndata: %d\n\n", time.Now().UnixMilli()); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
				// 心跳写出成功：连接存活，刷新在线活跃时间；写不出则上面已断开并按离线处理
				servants.NotifyServant.Heartbeat(uid)
			}
		}
	})
}

// PresenceHandler 批量查询用户在线状态（需登录）。
// query 参数：uids 逗号分隔的用户 id 列表（去重后 1~consts.NotifyMaxPresenceQueryUids 个）。
// 在线判定基于 SSE 连接心跳：有活跃连接且最近心跳/事件写出正常 → 在线，
// 否则离线（LastSeen 保留最近活跃时间，供展示「x 分钟前在线」）。
func PresenceHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if _, ok := middleware.CurrentUID(ctx); !ok {
		return resp.Unauthorized(ctx)
	}

	uids, ok := parseUIDs(ctx.Query("uids"))
	if !ok {
		return resp.BadRequest(ctx)
	}
	return resp.PresenceOK(ctx, servants.NotifyServant.Presence(uids))
}

// parseUIDs 解析逗号分隔的用户 id 列表（1~consts.NotifyMaxPresenceQueryUids 个，去重）。
func parseUIDs(raw string) ([]uint32, bool) {
	fields := strings.Split(raw, ",")
	uids := make([]uint32, 0, len(fields))
	seen := map[uint32]struct{}{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		id, err := strconv.ParseUint(field, 10, 32)
		if err != nil || id == 0 {
			return nil, false
		}
		uid := uint32(id)
		if _, dup := seen[uid]; dup {
			continue
		}
		seen[uid] = struct{}{}
		uids = append(uids, uid)
	}
	if len(uids) == 0 || len(uids) > consts.NotifyMaxPresenceQueryUids {
		return nil, false
	}
	return uids, true
}

// parseNotifyID 解析路径中的通知 id。
func parseNotifyID(ctx fiber.Ctx) (uint64, bool) {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}
