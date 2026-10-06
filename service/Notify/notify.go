package Notify

import (
	"context"
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
	"miaoverse/consts"
	daouser "miaoverse/dao/user"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
)

// Servant 通知业务服务：通知落库、查询/已读/删除编排、SSE 实时推送与在线状态。
// 组合用户 DAO（notify 表访问）、本实例 SSE 连接注册表（Hub），
// 以及 Redis 承载的跨实例状态（PresenceStore 在线状态、EventBus 事件总线），
// 支撑无状态化多实例部署。
type Servant struct {
	users    *daouser.UserDAO
	hub      *Hub
	presence *PresenceStore
	bus      *EventBus
}

// NewServant 创建通知服务。redisClient 用于在线状态与跨实例事件总线（cache db）。
func NewServant(users *daouser.UserDAO, redisClient *redis.Client) *Servant {
	return &Servant{
		users:    users,
		hub:      NewHub(),
		presence: NewPresenceStore(redisClient),
		bus:      NewEventBus(redisClient),
	}
}

// StartEventBus 启动跨实例通知事件总线接收（进程启动时调用一次；ctx 取消后停止）。
func (s *Servant) StartEventBus(ctx context.Context) {
	s.bus.Start(ctx, s.dispatchBusMessage)
}

// dispatchBusMessage 把其他实例广播的通知事件投递到本实例的 SSE 连接（送达则记录送达标记）。
func (s *Servant) dispatchBusMessage(msg BusMessage) {
	if s.hub.Publish(msg.UID, msg.Event) > 0 {
		if err := s.users.MarkNotifyReceived(msg.Event.Notify.ID); err != nil {
			log.Printf("Notify.Bus: 标记送达失败 notify=%d: %v", msg.Event.Notify.ID, err)
		}
	}
}

// Subscribe 订阅用户的通知 SSE 事件流，返回订阅句柄（Events 事件通道、Close 退订）。
// 连接心跳/事件成功写出后调用 Heartbeat 维持在线状态。
func (s *Servant) Subscribe(userID uint32) *Subscription {
	return s.hub.Subscribe(userID)
}

// Heartbeat SSE 连接心跳/事件成功写出后调用，刷新用户在线活跃时间（Redis，跨实例共享）。
func (s *Servant) Heartbeat(userID uint32) {
	s.presence.Touch(userID)
}

// Presence 批量查询用户在线状态（基于 SSE 连接心跳，Redis 跨实例共享，判定规则见 PresenceStore）。
// 返回结果按入参顺序；从未连接过的用户 Online=false、LastSeen 为零值。
func (s *Servant) Presence(userIDs []uint32) []resp.PresenceInfo {
	batch := s.presence.Presence(userIDs)
	items := make([]resp.PresenceInfo, 0, len(userIDs))
	for _, uid := range userIDs {
		p := batch[uid]
		items = append(items, resp.PresenceInfo{
			UID:      uid,
			Online:   p.Online,
			LastSeen: p.LastSeen,
		})
	}
	return items
}

// Event 通知事件入参。
type Event struct {
	Type       uint8           // 通知类型，取值见 consts.NotifyType*
	ReceiverID uint32          // 通知接收者
	Actor      *modeluser.User // 触发者用户，账号安全类通知（登录/改密）传 nil
	TargetType uint8           // 关联对象类型，复用 consts.InteractTarget*，无关联对象传 0
	TargetID   uint64          // 关联对象 id（用户 id / 动态 id / 评论 id）
	Content    string          // 摘要文本（如评论内容），超长自动截断
	Lang       string          // 触发请求的语言，用于已注销账号打码文案渲染
}

// Send 发送通知：写入 notify 表并通过 SSE 推送给接收者。
//
// 自发通知（Actor.ID == ReceiverID）不落库不推送，直接返回；
// 落库失败返回 error，由调用方决定是否上报（通知属旁路业务，通常只记录日志）；
// SSE 推送失败不影响本次发送结果：数据已入库，前端可拉取列表补齐。
func (s *Servant) Send(ev Event) (modeluser.Notify, error) {
	if ev.ReceiverID == 0 {
		return modeluser.Notify{}, nil
	}
	if ev.Actor != nil && ev.Actor.ID == ev.ReceiverID {
		return modeluser.Notify{}, nil
	}

	n := modeluser.Notify{
		UserID:     ev.ReceiverID,
		Type:       ev.Type,
		TargetType: ev.TargetType,
		TargetID:   ev.TargetID,
		Content:    truncateContent(ev.Content),
		Status:     consts.NotifyStatusUnread,
	}
	if ev.Actor != nil {
		n.ActorID = ev.Actor.ID
	}
	if err := s.users.CreateNotify(&n); err != nil {
		return modeluser.Notify{}, err
	}

	unread, err := s.users.CountUnreadNotifiesByTypes(ev.ReceiverID)
	if err != nil {
		log.Printf("Notify.Send: 统计未读数失败 user=%d: %v", ev.ReceiverID, err)
	}
	event := resp.NotifyEvent{
		Notify: ToNotifyInfo(n, ev.Actor, ev.Lang),
		Unread: ToUnreadCounts(unread),
	}
	// 先直接投递给本实例的 SSE 连接，再经事件总线广播给其他实例（无状态化部署下
	// 用户的连接可能在别的实例上，由总线消息触发跨实例投递；本实例发布的消息会被跳过）
	if s.hub.Publish(ev.ReceiverID, event) > 0 {
		if err := s.users.MarkNotifyReceived(n.ID); err != nil {
			log.Printf("Notify.Send: 标记送达失败 notify=%d: %v", n.ID, err)
		}
	}
	s.bus.Publish(ev.ReceiverID, event)
	return n, nil
}

// 以下为各业务触发点的语义化封装：约定好通知类型与关联对象，handler 只需传业务参数。
// 全部为旁路尽力而为：失败只记录日志，不影响点赞/评论/关注等主流程结果。

// NotifyLikeMoment 被点赞动态：通知动态作者（自己给自己点赞不通知）。
func (s *Servant) NotifyLikeMoment(actor *modeluser.User, momentID uint64, momentAuthor uint32, momentContent, lang string) {
	s.bestEffort(Event{
		Type:       consts.NotifyTypeLike,
		ReceiverID: momentAuthor,
		Actor:      actor,
		TargetType: consts.InteractTargetMoment,
		TargetID:   momentID,
		Content:    momentContent,
		Lang:       lang,
	})
}

// NotifyLikeComment 被点赞评论：通知评论作者（自己给自己的评论点赞不通知）。
func (s *Servant) NotifyLikeComment(actor *modeluser.User, commentID uint64, commentAuthor uint32, commentContent, lang string) {
	s.bestEffort(Event{
		Type:       consts.NotifyTypeLike,
		ReceiverID: commentAuthor,
		Actor:      actor,
		TargetType: consts.InteractTargetComment,
		TargetID:   commentID,
		Content:    commentContent,
		Lang:       lang,
	})
}

// NotifyComment 被评论动态：通知动态作者（自己评论自己的动态不通知）。
// Content 为评论内容摘要，Target 指向被评论的动态。
func (s *Servant) NotifyComment(actor *modeluser.User, momentID uint64, momentAuthor uint32, commentContent, lang string) {
	s.bestEffort(Event{
		Type:       consts.NotifyTypeReply,
		ReceiverID: momentAuthor,
		Actor:      actor,
		TargetType: consts.InteractTargetMoment,
		TargetID:   momentID,
		Content:    commentContent,
		Lang:       lang,
	})
}

// NotifyReply 被回复评论：通知被回复的评论作者（自己回复自己不通知）。
// Content 为回复内容摘要，Target 指向被回复的评论（前端据此展示「回复了你的评论」）。
func (s *Servant) NotifyReply(actor *modeluser.User, repliedID uint64, repliedAuthor uint32, replyContent, lang string) {
	s.bestEffort(Event{
		Type:       consts.NotifyTypeReply,
		ReceiverID: repliedAuthor,
		Actor:      actor,
		TargetType: consts.InteractTargetComment,
		TargetID:   repliedID,
		Content:    replyContent,
		Lang:       lang,
	})
}

// NotifyFollow 被关注：通知目标用户（关注自己不通知）。Target 指向关注者本人。
func (s *Servant) NotifyFollow(actor *modeluser.User, targetID uint32, lang string) {
	if actor == nil {
		return
	}
	s.bestEffort(Event{
		Type:       consts.NotifyTypeFollow,
		ReceiverID: targetID,
		Actor:      actor,
		TargetType: consts.InteractTargetUser,
		TargetID:   uint64(actor.ID),
		Lang:       lang,
	})
}

// NotifyAccountSecurity 账号安全通知（登录、修改密码）：通知账号本人。
// content 为已渲染的系统文案（由调用方按请求语言取 i18n 文案）。
func (s *Servant) NotifyAccountSecurity(uid uint32, content string) {
	s.bestEffort(Event{
		Type:       consts.NotifyTypeAccountSecurity,
		ReceiverID: uid,
		Content:    content,
	})
}

// bestEffort 尽力而为发送通知：失败只记录日志，不向调用方返回错误。
func (s *Servant) bestEffort(ev Event) {
	if _, err := s.Send(ev); err != nil {
		log.Printf("Notify: 通知发送失败 receiver=%d type=%d: %v", ev.ReceiverID, ev.Type, err)
	}
}

// List 分页查询用户通知（id 倒序）。
// category 为空字符串表示全部分类；取值非法时第二个返回值为 false。
func (s *Servant) List(userID uint32, category string, offset, limit int, lang string) ([]resp.NotifyInfo, int64, bool, error) {
	types, ok := TypesOfCategory(category)
	if !ok {
		return nil, 0, false, nil
	}

	notifies, err := s.users.QueryNotifiesByUser(userID, types, offset, limit)
	if err != nil {
		return nil, 0, false, err
	}
	count, err := s.users.CountNotifiesByUser(userID, types)
	if err != nil {
		return nil, 0, false, err
	}

	// 批量装配触发者，避免逐条查询
	actorIDs := make([]uint32, 0, len(notifies))
	for i := range notifies {
		if notifies[i].ActorID != 0 {
			actorIDs = append(actorIDs, notifies[i].ActorID)
		}
	}
	actors, err := s.users.QueryUsersByIDs(actorIDs)
	if err != nil {
		return nil, 0, false, err
	}

	items := make([]resp.NotifyInfo, 0, len(notifies))
	for i := range notifies {
		var actor *modeluser.User
		if notifies[i].ActorID != 0 {
			if u, found := actors[notifies[i].ActorID]; found {
				actor = &u
			}
		}
		items = append(items, ToNotifyInfo(notifies[i], actor, lang))
	}
	return items, count, true, nil
}

// Unread 查询用户各分类未读数。
func (s *Servant) Unread(userID uint32) (resp.NotifyUnread, error) {
	counts, err := s.users.CountUnreadNotifiesByTypes(userID)
	if err != nil {
		return resp.NotifyUnread{}, err
	}
	return ToUnreadCounts(counts), nil
}

// MarkRead 标记单条通知已读并返回最新未读数。
func (s *Servant) MarkRead(userID uint32, id uint64) (resp.NotifyUnread, error) {
	if err := s.users.MarkNotifyRead(id, userID); err != nil {
		return resp.NotifyUnread{}, err
	}
	return s.Unread(userID)
}

// MarkAllRead 全部标记已读并返回最新未读数。
func (s *Servant) MarkAllRead(userID uint32) (resp.NotifyUnread, error) {
	if err := s.users.MarkAllNotifyRead(userID); err != nil {
		return resp.NotifyUnread{}, err
	}
	return s.Unread(userID)
}

// Delete 删除单条通知（软删除）并返回最新未读数。
func (s *Servant) Delete(userID uint32, id uint64) (resp.NotifyUnread, error) {
	if err := s.users.DeleteNotify(id, userID); err != nil {
		return resp.NotifyUnread{}, err
	}
	return s.Unread(userID)
}

// CategoryOf 通知类型 → 前端分类（consts.NotifyCategory*）。
func CategoryOf(t uint8) string {
	switch t {
	case consts.NotifyTypeAccountSecurity, consts.NotifyTypeTransaction:
		return consts.NotifyCategoryAccount
	case consts.NotifyTypeLike:
		return consts.NotifyCategoryLike
	case consts.NotifyTypeFollow:
		return consts.NotifyCategoryFollow
	case consts.NotifyTypeMention:
		return consts.NotifyCategoryMention
	case consts.NotifyTypeReply:
		return consts.NotifyCategoryReply
	default:
		return consts.NotifyCategoryAccount
	}
}

// TypesOfCategory 前端分类 → 通知类型白名单。
// category 为空字符串表示全部分类（返回 nil）；未知分类返回 false。
func TypesOfCategory(category string) ([]uint8, bool) {
	switch strings.TrimSpace(category) {
	case "":
		return nil, true
	case consts.NotifyCategoryAccount:
		return []uint8{consts.NotifyTypeAccountSecurity, consts.NotifyTypeTransaction}, true
	case consts.NotifyCategoryLike:
		return []uint8{consts.NotifyTypeLike}, true
	case consts.NotifyCategoryFollow:
		return []uint8{consts.NotifyTypeFollow}, true
	case consts.NotifyCategoryMention:
		return []uint8{consts.NotifyTypeMention}, true
	case consts.NotifyCategoryReply:
		return []uint8{consts.NotifyTypeReply}, true
	default:
		return nil, false
	}
}

// ToNotifyInfo DAO 模型 → 响应 DTO（含已注销触发者打码，lang 为渲染语言）。
func ToNotifyInfo(n modeluser.Notify, actor *modeluser.User, lang string) resp.NotifyInfo {
	var actorCopy *modeluser.User
	if actor != nil {
		u := *actor
		resp.MaskClosedAccountByLang(lang, &u)
		actorCopy = &u
	}
	return resp.NotifyInfo{
		ID:         n.ID,
		Type:       n.Type,
		Category:   CategoryOf(n.Type),
		Actor:      actorCopy,
		TargetType: n.TargetType,
		TargetID:   n.TargetID,
		Content:    n.Content,
		CreatedAt:  n.CreatedAt,
		ReadAt:     n.ReadAt,
		Read:       n.Status == consts.NotifyStatusRead,
	}
}

// ToUnreadCounts 按类型分组的未读数 → 按分类分组的未读数。
func ToUnreadCounts(counts map[uint8]int64) resp.NotifyUnread {
	unread := resp.NotifyUnread{}
	for t, c := range counts {
		unread.Total += c
		switch CategoryOf(t) {
		case consts.NotifyCategoryAccount:
			unread.Account += c
		case consts.NotifyCategoryLike:
			unread.Like += c
		case consts.NotifyCategoryFollow:
			unread.Follow += c
		case consts.NotifyCategoryMention:
			unread.Mention += c
		case consts.NotifyCategoryReply:
			unread.Reply += c
		}
	}
	return unread
}

// truncateContent 按 rune 截断摘要，避免多字节字符被截断产生乱码。
func truncateContent(content string) string {
	content = strings.TrimSpace(content)
	runes := []rune(content)
	if len(runes) <= consts.MaxNotifyContentLen {
		return content
	}
	return string(runes[:consts.MaxNotifyContentLen])
}
