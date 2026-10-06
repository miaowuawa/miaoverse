package Notify

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"miaoverse/consts"
)

// Presence 用户在线状态。
type Presence struct {
	Online   bool      // 是否在线（最近心跳正常）
	LastSeen time.Time // 最近活跃时间，无记录为零值
}

// PresenceStore 用户在线状态存储（Redis ZSET，跨实例共享，支撑无状态化部署）。
// member=用户 id，score=最近一次心跳/事件成功写出的 Unix 毫秒时间戳。
//
// 在线判定完全以 SSE 连接心跳为准：距当前不超过 consts.NotifySSEOfflineAfter 即在线
// （心跳能收到就是在线）；连接断开/心跳写失败后无新写入，超过窗口自动转为离线
// （收不到心跳即判定离线），无需显式删除。
//
// 任何实例写入的心跳对所有实例可见，因此用户无论连到哪个实例，
// 查询接口都能拿到一致的在线状态。score 即最近活跃时间，离线后保留供展示
// 「x 分钟前在线」，超过 consts.NotifyPresenceRetention 由 Touch 机会性清理。
type PresenceStore struct {
	redis     *redis.Client
	key       string
	mu        sync.Mutex
	lastPrune time.Time
}

// NewPresenceStore 创建在线状态存储。
func NewPresenceStore(redisClient *redis.Client) *PresenceStore {
	return &PresenceStore{redis: redisClient, key: consts.NotifyPresenceKey}
}

// Touch 刷新用户活跃时间（心跳/事件成功写出后调用）。
func (p *PresenceStore) Touch(userID uint32) {
	if p == nil || p.redis == nil || userID == 0 {
		return
	}
	now := time.Now()
	_ = p.redis.ZAdd(context.Background(), p.key, &redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: strconv.FormatUint(uint64(userID), 10),
	}).Err()
	p.maybePrune(now)
}

// Presence 批量查询在线状态（结果覆盖全部入参用户，无记录的 Online=false、LastSeen 为零值）。
func (p *PresenceStore) Presence(userIDs []uint32) map[uint32]Presence {
	result := make(map[uint32]Presence, len(userIDs))
	if p == nil || p.redis == nil {
		return result
	}

	ctx := context.Background()
	now := time.Now()

	// 流水线批量 ZSCORE，一次网络往返
	pipe := p.redis.Pipeline()
	cmds := make(map[uint32]*redis.FloatCmd, len(userIDs))
	for _, userID := range userIDs {
		if userID == 0 {
			continue
		}
		if _, dup := cmds[userID]; dup {
			continue
		}
		cmds[userID] = pipe.ZScore(ctx, p.key, strconv.FormatUint(uint64(userID), 10))
	}
	_, _ = pipe.Exec(ctx)

	for userID, cmd := range cmds {
		score, err := cmd.Result()
		if err != nil {
			// redis.Nil：无记录 → 离线；其他错误同样按无记录处理（在线状态是尽力而为的旁路信息）
			continue
		}
		lastSeen := time.UnixMilli(int64(score))
		result[userID] = Presence{
			Online:   now.Sub(lastSeen) <= consts.NotifySSEOfflineAfter,
			LastSeen: lastSeen,
		}
	}
	return result
}

// maybePrune 机会性清理超过保留期的记录（限制触发频率，避免每次 Touch 都扫描）。
func (p *PresenceStore) maybePrune(now time.Time) {
	p.mu.Lock()
	if now.Sub(p.lastPrune) < time.Minute {
		p.mu.Unlock()
		return
	}
	p.lastPrune = now
	p.mu.Unlock()

	cutoff := strconv.FormatInt(now.Add(-consts.NotifyPresenceRetention).UnixMilli(), 10)
	_ = p.redis.ZRemRangeByScore(context.Background(), p.key, "-inf", cutoff).Err()
}
