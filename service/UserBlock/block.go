package UserBlock

import (
	"context"
	"errors"
	"fmt"

	"github.com/RoaringBitmap/roaring"
	"github.com/go-redis/redis/v8"
)

// BlockType 拉黑/屏蔽/不想看 三种关系类型
type BlockType uint8

const (
	BlockTypeBlock   BlockType = 1 // 拉黑
	BlockTypeMute    BlockType = 2 // 屏蔽
	BlockTypeUnwatch BlockType = 3 // 不想看
)

const (
	blockKeyPrefix = "block:user:"
	blockKeySuffix = ":"
)

var (
	ErrInvalidBlockType = errors.New("invalid block type")
	ErrInvalidUserID    = errors.New("invalid user id")
)

// Servant 用户拉黑/屏蔽/不想看 Bitmap 服务。
// 每个用户每种关系类型对应一个 Redis String，值为 RoaringBitmap 序列化后的字节数组（二进制安全）。
type Servant struct {
	redis *redis.Client
	db    int
}

func NewServant(redisClient *redis.Client, db int) *Servant {
	return &Servant{redis: redisClient, db: db}
}

// BlockKey 构造 Redis key：block:user:{userId}:{type}
func (s *Servant) BlockKey(userID uint32, blockType BlockType) (string, error) {
	if userID == 0 {
		return "", ErrInvalidUserID
	}
	if !validBlockType(blockType) {
		return "", ErrInvalidBlockType
	}
	return fmt.Sprintf("%s%d%s%d", blockKeyPrefix, userID, blockKeySuffix, blockType), nil
}

// Add 将 targetUserID 加入 userID 的指定关系 Bitmap。
// 用户第一次操作时自动创建空 Bitmap 并序列化存入 Redis。
func (s *Servant) Add(ctx context.Context, userID uint32, blockType BlockType, targetUserID uint32) error {
	if targetUserID == 0 {
		return ErrInvalidUserID
	}
	key, err := s.BlockKey(userID, blockType)
	if err != nil {
		return err
	}

	bitmap, err := s.load(ctx, key)
	if err != nil {
		return err
	}
	bitmap.Add(targetUserID)
	return s.save(ctx, key, bitmap)
}

// Remove 将 targetUserID 从 userID 的指定关系 Bitmap 中移除。
func (s *Servant) Remove(ctx context.Context, userID uint32, blockType BlockType, targetUserID uint32) error {
	if targetUserID == 0 {
		return ErrInvalidUserID
	}
	key, err := s.BlockKey(userID, blockType)
	if err != nil {
		return err
	}

	bitmap, err := s.load(ctx, key)
	if err != nil {
		return err
	}
	bitmap.Remove(targetUserID)
	return s.save(ctx, key, bitmap)
}

// Contains 查询 userID 的指定关系 Bitmap 中是否包含 targetUserID。
func (s *Servant) Contains(ctx context.Context, userID uint32, blockType BlockType, targetUserID uint32) (bool, error) {
	if targetUserID == 0 {
		return false, ErrInvalidUserID
	}
	key, err := s.BlockKey(userID, blockType)
	if err != nil {
		return false, err
	}

	bitmap, err := s.load(ctx, key)
	if err != nil {
		return false, err
	}
	return bitmap.Contains(targetUserID), nil
}

// Count 返回 userID 指定关系 Bitmap 中的用户数量。
func (s *Servant) Count(ctx context.Context, userID uint32, blockType BlockType) (uint64, error) {
	key, err := s.BlockKey(userID, blockType)
	if err != nil {
		return 0, err
	}

	bitmap, err := s.load(ctx, key)
	if err != nil {
		return 0, err
	}
	return bitmap.GetCardinality(), nil
}

// List 返回 userID 指定关系 Bitmap 中的全部用户 ID。
func (s *Servant) List(ctx context.Context, userID uint32, blockType BlockType) ([]uint32, error) {
	key, err := s.BlockKey(userID, blockType)
	if err != nil {
		return nil, err
	}

	bitmap, err := s.load(ctx, key)
	if err != nil {
		return nil, err
	}
	return bitmap.ToArray(), nil
}

// load 读取并反序列化 Bitmap；key 不存在时返回空 Bitmap（不写入 Redis）。
func (s *Servant) load(ctx context.Context, key string) (*roaring.Bitmap, error) {
	raw, err := s.redis.WithContext(ctx).Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return roaring.New(), nil
		}
		return nil, fmt.Errorf("read block bitmap %q: %w", key, err)
	}

	bitmap := roaring.New()
	if _, err := bitmap.FromBuffer(raw); err != nil {
		return nil, fmt.Errorf("deserialize block bitmap %q: %w", key, err)
	}
	return bitmap, nil
}

// save 序列化 Bitmap 并写入 Redis（二进制安全）。
func (s *Servant) save(ctx context.Context, key string, bitmap *roaring.Bitmap) error {
	raw, err := bitmap.ToBytes()
	if err != nil {
		return fmt.Errorf("serialize block bitmap %q: %w", key, err)
	}
	if err := s.redis.WithContext(ctx).Set(ctx, key, raw, 0).Err(); err != nil {
		return fmt.Errorf("write block bitmap %q: %w", key, err)
	}
	return nil
}

func validBlockType(blockType BlockType) bool {
	return blockType == BlockTypeBlock || blockType == BlockTypeMute || blockType == BlockTypeUnwatch
}
