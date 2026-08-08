package MetaSync

import (
	"context"
	"log"
	"time"

	"miaoverse/consts"
	modelmoment "miaoverse/model/dao/moment"
	"miaoverse/model/server"
)

// Start 启动动态计数定期校准任务，保证 moment_meta 与实际数量同步。
// 采用增量校准：只重算最近 consts.MetaSyncReconcileWindow 内有更新的动态，分批聚合统计后批量 upsert。
// 定时任务由调用方（cmd 启动流程）负责关闭。
func Start(ctx context.Context, servants *server.Servants, interval time.Duration) {
	if servants == nil || servants.ContentServant == nil || servants.InteractsServant == nil {
		log.Println("[metasync] servants not ready, skip periodic sync")
		return
	}
	if interval <= 0 {
		interval = 10 * time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := ReconcileMomentMetas(ctx, servants); err != nil {
					log.Printf("[metasync] reconcile failed: %v", err)
				}
			}
		}
	}()
}

// ReconcileMomentMetas 增量校准动态计数。
// 只处理最近 consts.MetaSyncReconcileWindow 内更新过的动态，按 consts.MetaSyncBatchSize 分批：
// 每批用 2 条 GROUP BY 聚合查询拿到真实点赞/评论数，再单条 SQL 批量 upsert。
func ReconcileMomentMetas(ctx context.Context, servants *server.Servants) error {
	since := time.Now().Add(-consts.MetaSyncReconcileWindow)
	ids, err := servants.ContentServant.QueryRecentMomentIDs(since)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	for start := 0; start < len(ids); start += consts.MetaSyncBatchSize {
		end := start + consts.MetaSyncBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]

		likes, err := servants.InteractsServant.CountMomentLikesBatch(chunk)
		if err != nil {
			return err
		}
		comments, err := servants.InteractsServant.CountMomentCommentsBatch(chunk)
		if err != nil {
			return err
		}

		updates := make(map[uint64]modelmoment.MomentMetaData, len(chunk))
		for _, id := range chunk {
			updates[id] = modelmoment.MomentMetaData{
				MomentID:     id,
				LikeCount:    uint32(likes[id]),
				CommentCount: uint32(comments[id]),
			}
		}
		if err := servants.ContentServant.UpsertMomentMetaCounts(updates); err != nil {
			return err
		}
	}
	return nil
}
