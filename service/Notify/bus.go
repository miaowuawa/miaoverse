package Notify

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"miaoverse/consts"
	"miaoverse/model/dto/resp"
)

// BusMessage 跨实例通知事件总线消息。
// Origin 为发布实例 id：通知在本实例产生时已直接投递给本实例的 SSE 连接，
// 总线消息由其他实例消费，接收方跳过自己发布的消息避免重复推送。
type BusMessage struct {
	Origin string           `json:"origin"`
	UID    uint32           `json:"uid"`
	Event  resp.NotifyEvent `json:"event"`
}

// EventBus 通知事件总线（Redis pub/sub，跨实例共享，支撑无状态化部署）。
//
// 无状态化部署下，产生通知的请求与用户的 SSE 连接可能落在不同实例：
// 通知写库后经总线广播给全部实例，各实例把事件投递到本实例上的连接，
// 保证无论用户连到哪个实例都能实时收到通知。
type EventBus struct {
	redis   *redis.Client
	channel string
	origin  string
}

// NewEventBus 创建事件总线（origin 为本实例随机 id）。
func NewEventBus(redisClient *redis.Client) *EventBus {
	return &EventBus{
		redis:   redisClient,
		channel: consts.NotifyBusChannel,
		origin:  uuid.NewString(),
	}
}

// Publish 广播事件给其他实例（本实例连接由调用方直接投递）。
// 广播失败只记录日志：通知已入库，前端可拉取列表补齐。
func (b *EventBus) Publish(uid uint32, ev resp.NotifyEvent) {
	if b == nil || b.redis == nil {
		return
	}
	payload, err := json.Marshal(BusMessage{Origin: b.origin, UID: uid, Event: ev})
	if err != nil {
		log.Printf("Notify.Bus: 消息序列化失败 user=%d: %v", uid, err)
		return
	}
	if err := b.redis.Publish(context.Background(), b.channel, payload).Err(); err != nil {
		log.Printf("Notify.Bus: 消息广播失败 user=%d: %v", uid, err)
	}
}

// Start 在后台协程订阅总线并持续分发消息（进程启动时调用一次；ctx 取消后退出）。
// 订阅连接异常断开后按 1 秒间隔自动重连。
func (b *EventBus) Start(ctx context.Context, onMessage func(BusMessage)) {
	if b == nil || b.redis == nil {
		return
	}
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			if err := b.listen(ctx, onMessage); err != nil && ctx.Err() == nil {
				log.Printf("Notify.Bus: 订阅中断，1 秒后重连: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
}

// listen 单次订阅循环：返回即代表订阅失效（由 Start 负责退避重连）。
func (b *EventBus) listen(ctx context.Context, onMessage func(BusMessage)) error {
	pubsub := b.redis.Subscribe(ctx, b.channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return errors.New("pubsub channel closed")
			}
			var m BusMessage
			if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil {
				log.Printf("Notify.Bus: 消息反序列化失败: %v", err)
				continue
			}
			// 本实例发布的消息已直接投递过，跳过避免重复推送
			if m.Origin == b.origin {
				continue
			}
			onMessage(m)
		}
	}
}
