package UserBlock

import (
	"context"
	"errors"
	"fmt"

	"github.com/RoaringBitmap/roaring"
	"github.com/go-redis/redis/v8"
	"miaoverse/consts"
)

// BlockType 拉黑/屏蔽/不想看 三种关系类型
type BlockType uint8

const (
	BlockTypeBlock   BlockType = 1 // 拉黑
	BlockTypeMute    BlockType = 2 // 屏蔽
	BlockTypeUnwatch BlockType = 3 // 不想看
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
	return fmt.Sprintf("%s%d%s%d", consts.BlockKeyPrefix, userID, consts.BlockKeySuffix, blockType), nil
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

// IsBlockedEither 判断 a 与 b 之间是否存在任意一方的拉黑关系（a 拉黑 b 或 b 拉黑 a）。
func (s *Servant) IsBlockedEither(ctx context.Context, a uint32, b uint32) (bool, error) {
	blocked, err := s.Contains(ctx, a, BlockTypeBlock, b)
	if err != nil {
		return false, err
	}
	if blocked {
		return true, nil
	}
	return s.Contains(ctx, b, BlockTypeBlock, a)
}

// IsFilteredBatch 批量判断 viewer 与每个 target 之间是否存在"拉黑/屏蔽/不想看"关系。
// 返回 targetID → 是否被过滤 的 map；未命中任何关系的 target 不在 map 中。
// 关系存储为 RoaringBitmap 序列化字节串（见 load/save），因此这里用 pipeline 批量 GET
// 反序列化后做成员判断（一次网络往返），不能用 GETBIT 对序列化字节做位偏移查询。
// 过滤口径：viewer 拉黑/屏蔽/不想看 target，或 target 拉黑 viewer（双向拉黑）。
// viewer=0 表示未登录游客：游客没有任何关系记录，直接返回空结果，不视为错误
// （timeline 等免登录 feed 会以 uid=0 调用本方法）。
func (s *Servant) IsFilteredBatch(ctx context.Context, viewer uint32, targets []uint32) (map[uint32]bool, error) {
	result := map[uint32]bool{}
	if viewer == 0 || len(targets) == 0 {
		return result, nil
	}

	// 去重，避免重复 key 查询
	seen := make(map[uint32]bool, len(targets))
	unique := make([]uint32, 0, len(targets))
	for _, t := range targets {
		if t == 0 || t == viewer || seen[t] {
			continue
		}
		seen[t] = true
		unique = append(unique, t)
	}
	if len(unique) == 0 {
		return result, nil
	}

	// viewer 侧 3 种关系各一个 Bitmap；target 侧只有拉黑 Bitmap
	viewerKeys := make([]string, 0, 3)
	for _, bt := range []BlockType{BlockTypeBlock, BlockTypeMute, BlockTypeUnwatch} {
		key, err := s.BlockKey(viewer, bt)
		if err != nil {
			return nil, err
		}
		viewerKeys = append(viewerKeys, key)
	}
	targetKeys := make(map[uint32]string, len(unique))
	keys := make([]string, 0, len(viewerKeys)+len(unique))
	keys = append(keys, viewerKeys...)
	for _, t := range unique {
		key, err := s.BlockKey(t, BlockTypeBlock)
		if err != nil {
			return nil, err
		}
		targetKeys[t] = key
		keys = append(keys, key)
	}

	// pipeline 一次往返读取全部 Bitmap 原始字节
	pipe := s.redis.WithContext(ctx).Pipeline()
	cmds := make([]*redis.StringCmd, 0, len(keys))
	for _, key := range keys {
		cmds = append(cmds, pipe.Get(ctx, key))
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("batch block filter: %w", err)
	}

	// key 不存在（无关系）时 GET 返回 redis.Nil，按空 Bitmap 处理
	bitmaps := make(map[string]*roaring.Bitmap, len(keys))
	for i, cmd := range cmds {
		raw, err := cmd.Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			return nil, fmt.Errorf("batch block filter: %w", err)
		}
		bitmap := roaring.New()
		if _, err := bitmap.FromBuffer(raw); err != nil {
			return nil, fmt.Errorf("deserialize block bitmap %q: %w", keys[i], err)
		}
		bitmaps[keys[i]] = bitmap
	}

	// viewer 侧 Bitmap 存放被 viewer 拉黑/屏蔽/不想看的用户；target 侧拉黑 Bitmap 存放被 target 拉黑的用户
	for _, t := range unique {
		for _, key := range viewerKeys {
			if b := bitmaps[key]; b != nil && b.Contains(t) {
				result[t] = true
				break
			}
		}
		if result[t] {
			continue
		}
		if b := bitmaps[targetKeys[t]]; b != nil && b.Contains(viewer) {
			result[t] = true
		}
	}
	return result, nil
}

// GetBlockStatus 计算 userID 对 targetID 的关系状态（位组合）：
// bit0(1) 拉黑，bit1(2) 屏蔽，bit2(4) 不想看；0 表示无任何关系。
func (s *Servant) GetBlockStatus(ctx context.Context, userID uint32, targetID uint32) (uint8, error) {
	var status uint8
	for _, blockType := range []BlockType{BlockTypeBlock, BlockTypeMute, BlockTypeUnwatch} {
		contained, err := s.Contains(ctx, userID, blockType, targetID)
		if err != nil {
			return 0, err
		}
		if contained {
			status |= 1 << (blockType - 1)
		}
	}
	return status, nil
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
