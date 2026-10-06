package Notify

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"miaoverse/consts"
)

func newTestPresenceStore(t *testing.T) (*PresenceStore, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return NewPresenceStore(client), client
}

// TestPresenceByHeartbeat 在线状态以 SSE 心跳写入为准：
// 心跳正常 → 在线；心跳停止超过离线判定窗口 → 离线；离线后保留最近活跃时间。
func TestPresenceByHeartbeat(t *testing.T) {
	store, client := newTestPresenceStore(t)

	// 从未连接过
	p := store.Presence([]uint32{1})[1]
	if p.Online || !p.LastSeen.IsZero() {
		t.Fatalf("never touched: %+v", p)
	}

	// 心跳写入 → 在线
	store.Touch(1)
	p = store.Presence([]uint32{1})[1]
	if !p.Online {
		t.Fatalf("fresh heartbeat should be online: %+v", p)
	}
	fresh := p.LastSeen

	// 心跳停止（最近写入停留在离线窗口之前）→ 离线，保留最近活跃时间
	staleScore := float64(time.Now().Add(-consts.NotifySSEOfflineAfter - time.Second).UnixMilli())
	if err := client.ZAdd(context.Background(), consts.NotifyPresenceKey,
		&redis.Z{Score: staleScore, Member: "1"}).Err(); err != nil {
		t.Fatalf("ZAdd: %v", err)
	}
	p = store.Presence([]uint32{1})[1]
	if p.Online {
		t.Fatalf("stale heartbeat should be offline: %+v", p)
	}
	if p.LastSeen.IsZero() {
		t.Fatal("LastSeen should be retained when offline")
	}

	// 心跳恢复 → 重新在线
	store.Touch(1)
	p = store.Presence([]uint32{1})[1]
	if !p.Online {
		t.Fatalf("heartbeat resumed should be online: %+v", p)
	}
	if p.LastSeen.Before(fresh) {
		t.Fatal("LastSeen should be refreshed by Touch")
	}
}

// TestPresenceBatch 跨用户批量查询：结果覆盖全部入参用户，无记录/零 id 按离线处理。
func TestPresenceBatch(t *testing.T) {
	store, _ := newTestPresenceStore(t)
	store.Touch(2)

	batch := store.Presence([]uint32{2, 3, 0})
	if !batch[2].Online {
		t.Fatalf("uid 2 should be online: %+v", batch[2])
	}
	if batch[3].Online || !batch[3].LastSeen.IsZero() {
		t.Fatalf("uid 3 should be offline with zero LastSeen: %+v", batch[3])
	}
	if batch[0].Online {
		t.Fatalf("uid 0 should be zero value: %+v", batch[0])
	}
}

// TestPresencePrune 超过保留期的记录被清理（Touch 机会性触发）。
func TestPresencePrune(t *testing.T) {
	store, client := newTestPresenceStore(t)

	// 写入一条超过保留期的旧记录
	oldScore := float64(time.Now().Add(-consts.NotifyPresenceRetention - time.Minute).UnixMilli())
	if err := client.ZAdd(context.Background(), consts.NotifyPresenceKey,
		&redis.Z{Score: oldScore, Member: "6"}).Err(); err != nil {
		t.Fatalf("ZAdd: %v", err)
	}

	// Touch 触发清理
	store.Touch(5)
	if p := store.Presence([]uint32{6})[6]; !p.LastSeen.IsZero() {
		t.Fatalf("stale record should be pruned: %+v", p)
	}
	if p := store.Presence([]uint32{5})[5]; !p.Online {
		t.Fatalf("recent record should be kept: %+v", p)
	}
}

// TestPresenceNilGuards nil 依赖安全（测试与降级场景不 panic）。
func TestPresenceNilGuards(t *testing.T) {
	var nilStore *PresenceStore
	nilStore.Touch(1)
	if got := nilStore.Presence([]uint32{1}); len(got) != 0 {
		t.Fatalf("nil store Presence = %+v, want empty", got)
	}

	empty := NewPresenceStore(nil)
	empty.Touch(1)
	if got := empty.Presence([]uint32{1}); len(got) != 0 {
		t.Fatalf("nil redis Presence = %+v, want empty", got)
	}
}

// TestServantPresence Servant 在线状态查询（基于 Redis 心跳写入）。
func TestServantPresence(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	s := NewServant(nil, client)

	s.Heartbeat(7)

	items := s.Presence([]uint32{7, 8})
	if len(items) != 2 {
		t.Fatalf("Presence size = %d, want 2", len(items))
	}
	if items[0].UID != 7 || !items[0].Online {
		t.Fatalf("items[0] = %+v, want uid 7 online", items[0])
	}
	if items[1].UID != 8 || items[1].Online {
		t.Fatalf("items[1] = %+v, want uid 8 offline", items[1])
	}
}
