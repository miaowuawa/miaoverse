package Notify

import (
	"sync"

	"miaoverse/consts"
	"miaoverse/model/dto/resp"
)

// Hub 本实例的 SSE 通知推送连接注册表。
// 每个在线用户可有多个 SSE 连接（多标签页），每个连接对应一个独立事件通道。
//
// Hub 只持有本实例的连接（连接本身无法跨实例共享）；跨实例实时推送由 EventBus
// （Redis pub/sub）广播后各实例向本实例连接投递，在线状态由 PresenceStore（Redis）维护。
//
// Publish 为非阻塞投递：连接缓冲（consts.NotifySSEBufferSize）写满时丢弃该事件，
// 通知数据以数据库为准，前端可随时拉取列表补齐。
type Hub struct {
	mu     sync.RWMutex
	subs   map[uint32]map[chan resp.NotifyEvent]struct{}
	closed bool
}

// Subscription 一条 SSE 订阅连接的句柄。
type Subscription struct {
	// Events 通知事件通道；Hub 关闭或 Close() 后通道被关闭。
	Events <-chan resp.NotifyEvent
	close  func()
}

// Close 退订并结束连接（可重复调用）。
func (s *Subscription) Close() {
	s.close()
}

// NewHub 创建连接注册表。
func NewHub() *Hub {
	return &Hub{subs: map[uint32]map[chan resp.NotifyEvent]struct{}{}}
}

// Subscribe 订阅用户的通知事件流，返回订阅句柄。
// 连接关闭（Hub.Close）后订阅会立即结束：通道被关闭，Close 不再持有注册。
func (h *Hub) Subscribe(userID uint32) *Subscription {
	ch := make(chan resp.NotifyEvent, consts.NotifySSEBufferSize)

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		close(ch)
		return &Subscription{Events: ch, close: func() {}}
	}
	set, ok := h.subs[userID]
	if !ok {
		set = map[chan resp.NotifyEvent]struct{}{}
		h.subs[userID] = set
	}
	set[ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if set, ok := h.subs[userID]; ok {
				if _, ok := set[ch]; ok {
					delete(set, ch)
					if len(set) == 0 {
						delete(h.subs, userID)
					}
					close(ch)
				}
			}
		})
	}
	return &Subscription{Events: ch, close: unsubscribe}
}

// Publish 向用户的全部本地连接投递事件，返回成功投递的连接数。
// 非阻塞：缓冲写满的连接跳过该事件（该连接视为未送达）。
func (h *Hub) Publish(userID uint32, ev resp.NotifyEvent) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return 0
	}
	delivered := 0
	for ch := range h.subs[userID] {
		select {
		case ch <- ev:
			delivered++
		default:
		}
	}
	return delivered
}

// Close 关闭连接注册表并结束全部订阅通道（服务停机时调用，让各 SSE 连接立即返回）。
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for userID, set := range h.subs {
		for ch := range set {
			close(ch)
		}
		delete(h.subs, userID)
	}
}

// Len 用户当前本地连接数（测试与监控用）。
func (h *Hub) Len(userID uint32) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs[userID])
}
