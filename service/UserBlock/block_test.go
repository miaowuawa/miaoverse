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
