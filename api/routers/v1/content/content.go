package content

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/consts"
	"miaoverse/middleware"
	modelarticle "miaoverse/model/dao/article"
	modelmoment "miaoverse/model/dao/moment"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/Moment"
	"miaoverse/util/pagination"
)

func CountHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	targetID, ok := middleware.BlockTarget(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	isFriend, isFan, err := relationFlags(ctx, servants, uid, targetID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	count, err := servants.ContentServant.CountVisibleMomentsByUser(uid, targetID, isFriend, isFan)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.ContentCount(ctx, count)
}

// ListHandler 获取目标用户对当前登录用户可见的内容列表。
// category 参数：moment（动态，默认）/ article（文章）/ novel（小说）。
func ListHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	targetID, ok := middleware.BlockTarget(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	offset, limit, ok := parsePagination(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	category := strings.TrimSpace(ctx.Query("category"))
	if category == "" {
		category = consts.FeedContentMoment
	}
	if category != consts.FeedContentMoment && category != consts.FeedContentArticle && category != consts.FeedContentNovel {
		return resp.BadRequest(ctx)
	}

	isFriend, isFan, err := relationFlags(ctx, servants, uid, targetID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	switch category {
	case consts.FeedContentMoment:
		moments, err := servants.ContentServant.QueryVisibleMomentsByUser(uid, targetID, isFriend, isFan, offset, limit)
		if err != nil {
			return resp.ServerError(ctx)
		}
		metas, err := servants.ContentServant.QueryMomentInteractCounts(momentIDs(moments))
		if err != nil {
			return resp.ServerError(ctx)
		}
		items := make([]resp.ContentItem, 0, len(moments))
		for i := range moments {
			items = append(items, Moment.ToContentItem(&moments[i], metas))
		}
		return resp.ContentList(ctx, items)
	case consts.FeedContentNovel:
		metas, err := servants.ArticleServant.QueryUserNovels(targetID, consts.ArticleStatusNormal, offset, limit)
		if err != nil {
			return resp.ServerError(ctx)
		}
		items, err := articleContentItems(servants, metas)
		if err != nil {
			return resp.ServerError(ctx)
		}
		return resp.ContentList(ctx, items)
	default: // article
		metas, err := servants.ArticleServant.QueryUserMetadatasByType(targetID, consts.ArticleTypeNormal, consts.ArticleStatusNormal, offset, limit)
		if err != nil {
			return resp.ServerError(ctx)
		}
		items, err := articleContentItems(servants, metas)
		if err != nil {
			return resp.ServerError(ctx)
		}
		return resp.ContentList(ctx, items)
	}
}

// articleContentItems 文章元数据 → 内容列表项（点赞/评论计数；小说附章节数，章节数批量统计避免 N+1）。
func articleContentItems(servants *server.Servants, metas []modelarticle.Metadata) ([]resp.ContentItem, error) {
	ids := make([]uint64, 0, len(metas))
	for i := range metas {
		ids = append(ids, metas[i].ID)
	}
	counts, err := servants.ArticleServant.QueryInteractCounts(ids)
	if err != nil {
		return nil, err
	}

	novelIDs := make([]uint64, 0, len(metas))
	for i := range metas {
		if metas[i].Type == consts.ArticleTypeNovel && metas[i].NovelID == 0 {
			novelIDs = append(novelIDs, metas[i].ID)
		}
	}
	chapterCounts, err := servants.ArticleServant.CountChaptersByNovels(novelIDs)
	if err != nil {
		return nil, err
	}

	items := make([]resp.ContentItem, 0, len(metas))
	for i := range metas {
		m := &metas[i]
		item := resp.ContentItem{
			ID:           m.ID,
			Type:         consts.ContentTypeArticle,
			ChapterCount: chapterCounts[m.ID],
		}
		if c, ok := counts[m.ID]; ok {
			item.Comment = c.CommentCount
			item.Like = c.LikeCount
		}
		items = append(items, item)
	}
	return items, nil
}

func parsePagination(ctx fiber.Ctx) (int, int, bool) {
	return pagination.Parse(ctx.Query("offset"), ctx.Query("limit"))
}

// relationFlags 计算查看者与目标用户的关系：isFriend 互相关注，isFan 目标关注了查看者
func relationFlags(ctx fiber.Ctx, servants *server.Servants, viewerID uint32, targetID uint32) (bool, bool, error) {
	if viewerID == targetID {
		return false, false, nil
	}

	following, err := servants.InteractsServant.IsFollowing(viewerID, targetID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, false, err
	}
	followedBy, err := servants.InteractsServant.IsFollowing(targetID, viewerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, false, err
	}
	return following && followedBy, followedBy, nil
}

func momentIDs(moments []modelmoment.Moment) []uint64 {
	ids := make([]uint64, 0, len(moments))
	for i := range moments {
		ids = append(ids, moments[i].ID)
	}
	return ids
}
