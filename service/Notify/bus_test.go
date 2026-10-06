package Notify

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"miaoverse/model/dto/resp"
)

// TestBusCrossInstanceDelivery 事件总线跨实例分发（无状态化部署核心路径）：
// 发布实例直接投递本地连接并广播；其他实例经总线收到消息，
// 发布实例跳过自己发布的消息（避免重复推送）。
func TestBusCrossInstanceDelivery(t *testing.T) {
	mr := miniredis.RunT(t)
	clientA := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	clientB := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = clientA.Close()
		_ = clientB.Close()
	})

	busA := NewEventBus(clientA)
	busB := NewEventBus(clientB)
	if busA.origin == busB.origin {
		t.Fatal("each bus should have a unique instance id")
	}

	receivedA := make(chan BusMessage, 16)
	receivedB := make(chan BusMessage, 16)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	busA.Start(ctx, func(m BusMessage) { receivedA <- m })
	busB.Start(ctx, func(m BusMessage) { receivedB <- m })

	// 重试发布直至订阅建立（Start 的订阅在后台协程异步完成）
	ev := resp.NotifyEvent{Notify: resp.NotifyInfo{ID: 99}}
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				busA.Publish(7, ev)
			}
		}
	}()

	select {
	case m := <-receivedB:
		if m.UID != 7 || m.Event.Notify.ID != 99 {
			t.Fatalf("bus B got %+v, want uid 7 notify 99", m)
		}
		if m.Origin != busA.origin {
			t.Fatalf("message origin = %q, want %q", m.Origin, busA.origin)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("bus B did not receive the message")
	}

	// 发布方自己不消费总线消息（本地已直接投递）
	select {
	case m := <-receivedA:
		t.Fatalf("bus A should skip its own message, got %+v", m)
	case <-time.After(300 * time.Millisecond):
	}
}

// TestBusNilGuards nil 依赖安全（测试与降级场景不 panic）。
func TestBusNilGuards(t *testing.T) {
	var nilBus *EventBus
	nilBus.Publish(1, resp.NotifyEvent{})
	nilBus.Start(context.Background(), func(BusMessage) {})

	empty := NewEventBus(nil)
	empty.Publish(1, resp.NotifyEvent{})
	empty.Start(context.Background(), func(BusMessage) {})
}
