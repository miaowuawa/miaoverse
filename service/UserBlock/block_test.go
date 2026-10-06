package UserBlock

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func newTestServant(t *testing.T) (*Servant, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return NewServant(client, 0), mr
}

func TestAddContainsRemove(t *testing.T) {
	servant, mr := newTestServant(t)
	ctx := context.Background()

	ok, err := servant.Contains(ctx, 10001, BlockTypeBlock, 20002)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("empty bitmap should not contain target")
	}

	if err := servant.Add(ctx, 10001, BlockTypeBlock, 20002); err != nil {
		t.Fatal(err)
	}
	ok, err = servant.Contains(ctx, 10001, BlockTypeBlock, 20002)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("target should be contained after add")
	}

	key, _ := servant.BlockKey(10001, BlockTypeBlock)
	if !mr.Exists(key) {
		t.Fatalf("key %q should exist after first add", key)
	}

	if err := servant.Remove(ctx, 10001, BlockTypeBlock, 20002); err != nil {
		t.Fatal(err)
	}
	ok, err = servant.Contains(ctx, 10001, BlockTypeBlock, 20002)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("target should not be contained after remove")
	}
}

func TestTypesAreIsolated(t *testing.T) {
	servant, _ := newTestServant(t)
	ctx := context.Background()

	if err := servant.Add(ctx, 10001, BlockTypeBlock, 20002); err != nil {
		t.Fatal(err)
	}
	if err := servant.Add(ctx, 10001, BlockTypeMute, 20003); err != nil {
		t.Fatal(err)
	}

	ok, err := servant.Contains(ctx, 10001, BlockTypeBlock, 20003)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("block bitmap must not contain mute target")
	}
	ok, err = servant.Contains(ctx, 10001, BlockTypeMute, 20002)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("mute bitmap must not contain block target")
	}
}

func TestCountAndList(t *testing.T) {
	servant, _ := newTestServant(t)
	ctx := context.Background()

	for _, id := range []uint32{20001, 20002, 20003} {
		if err := servant.Add(ctx, 10001, BlockTypeUnwatch, id); err != nil {
			t.Fatal(err)
		}
	}

	count, err := servant.Count(ctx, 10001, BlockTypeUnwatch)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}

	list, err := servant.List(ctx, 10001, BlockTypeUnwatch)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("list len = %d, want 3", len(list))
	}
}

// TestIsFilteredBatchGuest 未登录游客（viewer=0）不携带任何关系记录：
// 必须返回空过滤结果且不报错（timeline 等免登录 feed 以 uid=0 调用，曾因此返回 500）。
func TestIsFilteredBatchGuest(t *testing.T) {
	servant, _ := newTestServant(t)
	ctx := context.Background()

	filtered, err := servant.IsFilteredBatch(ctx, 0, []uint32{20001, 20002})
	if err != nil {
		t.Fatalf("guest batch filter error = %v, want nil", err)
	}
	if len(filtered) != 0 {
		t.Fatalf("guest filtered = %v, want empty", filtered)
	}
}

// TestIsFilteredBatchViewerRelations viewer 拉黑/屏蔽/不想看的 target 被过滤，其余保留。
// 同时覆盖多 target 场景（viewer 侧 key 共用时不能互相串扰）。
func TestIsFilteredBatchViewerRelations(t *testing.T) {
	servant, _ := newTestServant(t)
	ctx := context.Background()

	if err := servant.Add(ctx, 10001, BlockTypeBlock, 20002); err != nil {
		t.Fatal(err)
	}
	if err := servant.Add(ctx, 10001, BlockTypeMute, 20003); err != nil {
		t.Fatal(err)
	}
	if err := servant.Add(ctx, 10001, BlockTypeUnwatch, 20005); err != nil {
		t.Fatal(err)
	}

	filtered, err := servant.IsFilteredBatch(ctx, 10001, []uint32{20001, 20002, 20003, 20004, 20005})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []uint32{20002, 20003, 20005} {
		if !filtered[id] {
			t.Fatalf("target %d should be filtered, got %v", id, filtered)
		}
	}
	for _, id := range []uint32{20001, 20004} {
		if filtered[id] {
			t.Fatalf("target %d should not be filtered, got %v", id, filtered)
		}
	}
}

// TestIsFilteredBatchTargetBlocksViewer 双向口径：target 拉黑 viewer 时该 target 也被过滤。
func TestIsFilteredBatchTargetBlocksViewer(t *testing.T) {
	servant, _ := newTestServant(t)
	ctx := context.Background()

	if err := servant.Add(ctx, 20002, BlockTypeBlock, 10001); err != nil {
		t.Fatal(err)
	}

	filtered, err := servant.IsFilteredBatch(ctx, 10001, []uint32{20002, 20003})
	if err != nil {
		t.Fatal(err)
	}
	if !filtered[20002] {
		t.Fatalf("target 20002 blocked viewer but not filtered, got %v", filtered)
	}
	if filtered[20003] {
		t.Fatalf("target 20003 should not be filtered, got %v", filtered)
	}
}

func TestInvalidInputs(t *testing.T) {
	servant, _ := newTestServant(t)
	ctx := context.Background()

	if err := servant.Add(ctx, 0, BlockTypeBlock, 20002); err != ErrInvalidUserID {
		t.Fatalf("Add with zero userID error = %v, want ErrInvalidUserID", err)
	}
	if err := servant.Add(ctx, 10001, BlockType(99), 20002); err != ErrInvalidBlockType {
		t.Fatalf("Add with bad type error = %v, want ErrInvalidBlockType", err)
	}
	if err := servant.Add(ctx, 10001, BlockTypeBlock, 0); err != ErrInvalidUserID {
		t.Fatalf("Add with zero target error = %v, want ErrInvalidUserID", err)
	}
}
