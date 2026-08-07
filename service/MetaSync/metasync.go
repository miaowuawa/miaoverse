package MetaSync

import (
	"context"
	"log"
	"time"

	modelmoment "miaoverse/model/dao/moment"
	"miaoverse/model/server"
)

// Start 启动动态计数定期校准任务，保证 moment_meta 与实际数量同步。
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

// ReconcileMomentMetas 全量校准所有动态的点赞数/评论数（覆盖式写入，幂等）。
func ReconcileMomentMetas(ctx context.Context, servants *server.Servants) error {
	moments, err := servants.ContentServant.QueryAllMomentIDs()
	if err != nil {
		return err
	}
	if len(moments) == 0 {
		return nil
	}

	updates := make(map[uint64]modelmoment.MomentMetaData, len(moments))
	for _, m := range moments {
		likes, err := servants.InteractsServant.CountMomentLikesReal(m.ID)
		if err != nil {
			return err
		}
		comments, err := servants.InteractsServant.CountMomentCommentsReal(m.ID)
		if err != nil {
			return err
		}
		updates[m.ID] = modelmoment.MomentMetaData{
			MomentID:     m.ID,
			LikeCount:    uint32(likes),
			CommentCount: uint32(comments),
		}
	}

	return servants.ContentServant.SetMomentMetaCounts(updates)
}
