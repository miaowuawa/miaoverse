package Notify

import (
	"testing"
	"time"

	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
)

func TestHubPublishDeliversToAllConnections(t *testing.T) {
	h := NewHub()
	sub1 := h.Subscribe(1)
	sub2 := h.Subscribe(1)
	sub3 := h.Subscribe(2)
	defer sub1.Close()
	defer sub2.Close()
	defer sub3.Close()

	if got := h.Len(1); got != 2 {
		t.Fatalf("Len(1) = %d, want 2", got)
	}

	ev := resp.NotifyEvent{Notify: resp.NotifyInfo{ID: 42}}
	if got := h.Publish(1, ev); got != 2 {
		t.Fatalf("Publish delivered = %d, want 2", got)
	}
	if got := h.Publish(2, ev); got != 1 {
		t.Fatalf("Publish to user 2 delivered = %d, want 1", got)
	}

	for i, sub := range []*Subscription{sub1, sub2, sub3} {
		select {
		case got := <-sub.Events:
			if got.Notify.ID != 42 {
				t.Fatalf("conn %d got event id = %d, want 42", i, got.Notify.ID)
			}
		case <-time.After(time.Second):
			t.Fatalf("conn %d did not receive event", i)
		}
	}
}

func TestHubCloseSubscriptionStopsDelivery(t *testing.T) {
	h := NewHub()
	sub := h.Subscribe(1)

	sub.Close()
	sub.Close() // 重复关闭应安全

	if got := h.Len(1); got != 0 {
		t.Fatalf("Len(1) after close = %d, want 0", got)
	}
	if got := h.Publish(1, resp.NotifyEvent{}); got != 0 {
		t.Fatalf("Publish delivered = %d, want 0", got)
	}
	if _, ok := <-sub.Events; ok {
		t.Fatal("channel should be closed after subscription close")
	}
}

func TestHubPublishDropsWhenBufferFull(t *testing.T) {
	h := NewHub()
	sub := h.Subscribe(1)
	defer sub.Close()

	// 不消费通道，写满缓冲后（consts.NotifySSEBufferSize）后续事件被丢弃
	delivered := 0
	for i := 0; i < consts.NotifySSEBufferSize+5; i++ {
		delivered += h.Publish(1, resp.NotifyEvent{})
	}
	if delivered != consts.NotifySSEBufferSize {
		t.Fatalf("delivered = %d, want %d", delivered, consts.NotifySSEBufferSize)
	}
}

func TestHubCloseEndsAllSubscriptions(t *testing.T) {
	h := NewHub()
	sub1 := h.Subscribe(1)
	sub2 := h.Subscribe(2)

	h.Close()
	h.Close() // 幂等

	for i, sub := range []*Subscription{sub1, sub2} {
		if _, ok := <-sub.Events; ok {
			t.Fatalf("conn %d channel should be closed after Close", i)
		}
	}
	if got := h.Publish(1, resp.NotifyEvent{}); got != 0 {
		t.Fatalf("Publish after Close delivered = %d, want 0", got)
	}
	// 关闭后新订阅立即返回已关闭通道
	sub := h.Subscribe(1)
	defer sub.Close()
	if _, ok := <-sub.Events; ok {
		t.Fatal("subscribe after Close should return closed channel")
	}
}

func TestSendSkipsSelfNotification(t *testing.T) {
	// 自发通知在触达 DAO/Hub/Redis 前即返回，nil 依赖也不会 panic
	s := NewServant(nil, nil)
	actor := &modeluser.User{ID: 7}
	n, err := s.Send(Event{Type: consts.NotifyTypeLike, ReceiverID: 7, Actor: actor})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if n.ID != 0 {
		t.Fatalf("self notification should be skipped, got id = %d", n.ID)
	}

	// ReceiverID 为 0 同样跳过
	if _, err := s.Send(Event{Type: consts.NotifyTypeLike}); err != nil {
		t.Fatalf("Send() with zero receiver error = %v", err)
	}
}

func TestCategoryMapping(t *testing.T) {
	cases := []struct {
		typ      uint8
		category string
	}{
		{consts.NotifyTypeAccountSecurity, consts.NotifyCategoryAccount},
		{consts.NotifyTypeTransaction, consts.NotifyCategoryAccount},
		{consts.NotifyTypeLike, consts.NotifyCategoryLike},
		{consts.NotifyTypeFollow, consts.NotifyCategoryFollow},
		{consts.NotifyTypeMention, consts.NotifyCategoryMention},
		{consts.NotifyTypeReply, consts.NotifyCategoryReply},
	}
	for _, c := range cases {
		if got := CategoryOf(c.typ); got != c.category {
			t.Errorf("CategoryOf(%d) = %q, want %q", c.typ, got, c.category)
		}
		types, ok := TypesOfCategory(c.category)
		if !ok {
			t.Fatalf("TypesOfCategory(%q) not ok", c.category)
		}
		found := false
		for _, typ := range types {
			if typ == c.typ {
				found = true
			}
		}
		if !found {
			t.Errorf("TypesOfCategory(%q) = %v does not contain type %d", c.category, types, c.typ)
		}
	}

	if types, ok := TypesOfCategory(""); !ok || types != nil {
		t.Errorf("TypesOfCategory(\"\") = %v, %v; want nil, true", types, ok)
	}
	if _, ok := TypesOfCategory("bogus"); ok {
		t.Error("TypesOfCategory(\"bogus\") should not be ok")
	}
}

func TestToUnreadCounts(t *testing.T) {
	unread := ToUnreadCounts(map[uint8]int64{
		consts.NotifyTypeAccountSecurity: 1,
		consts.NotifyTypeTransaction:     2,
		consts.NotifyTypeLike:            3,
		consts.NotifyTypeFollow:          4,
		consts.NotifyTypeMention:         5,
		consts.NotifyTypeReply:           6,
	})
	want := resp.NotifyUnread{Total: 21, Account: 3, Like: 3, Follow: 4, Mention: 5, Reply: 6}
	if unread != want {
		t.Fatalf("ToUnreadCounts = %+v, want %+v", unread, want)
	}
}

func TestTruncateContent(t *testing.T) {
	if got := truncateContent("  你好  "); got != "你好" {
		t.Fatalf("trim failed, got %q", got)
	}
	long := make([]rune, consts.MaxNotifyContentLen+10)
	for i := range long {
		long[i] = '好'
	}
	got := truncateContent(string(long))
	if r := []rune(got); len(r) != consts.MaxNotifyContentLen {
		t.Fatalf("truncate length = %d, want %d", len(r), consts.MaxNotifyContentLen)
	}
}

func TestToNotifyInfo(t *testing.T) {
	now := time.Now()
	readAt := now.Add(time.Minute)
	n := modeluser.Notify{
		ID:         9,
		UserID:     1,
		Type:       consts.NotifyTypeReply,
		ActorID:    2,
		TargetType: consts.InteractTargetMoment,
		TargetID:   100,
		Content:    "内容",
		CreatedAt:  now,
		ReadAt:     &readAt,
		Status:     consts.NotifyStatusRead,
	}
	actor := &modeluser.User{ID: 2, Username: "actor", Nickname: "昵称"}
	info := ToNotifyInfo(n, actor, "zh-CN")

	if info.Category != consts.NotifyCategoryReply {
		t.Errorf("Category = %q, want %q", info.Category, consts.NotifyCategoryReply)
	}
	if !info.Read {
		t.Error("Read should be true for status=read")
	}
	if info.Actor == nil || info.Actor.ID != 2 {
		t.Fatalf("Actor = %+v, want id 2", info.Actor)
	}
	if info.Actor == actor {
		t.Error("ToNotifyInfo should copy actor, not alias it")
	}

	// 无触发者（账号安全类通知）
	info = ToNotifyInfo(modeluser.Notify{Type: consts.NotifyTypeAccountSecurity}, nil, "zh-CN")
	if info.Actor != nil {
		t.Fatalf("Actor = %+v, want nil", info.Actor)
	}
}
