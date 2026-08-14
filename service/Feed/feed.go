// Package Feed 提供 feed（信息流）组装逻辑：时间线（timeline）与关注流（following）。
//
// 设计要点：
//   - 动态与文章分别查询（不同表/不同库），在内存中按发布时间倒序合并；
//   - 屏蔽/删除等非正常状态在 DAO 层直接过滤，不进入 feed；
//   - 拉黑/屏蔽/不想看关系在 UserBlock 批量过滤（Redis pipeline 一次往返）；
//   - 作者、互动计数、点赞状态均批量查询，避免 N+1。
package Feed

import (
	"context"
	"errors"
	"sort"

	"gorm.io/gorm"
	"miaoverse/consts"
	modelarticle "miaoverse/model/dao/article"
	modelmoment "miaoverse/model/dao/moment"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
)

// QueryParams feed 查询参数。
type QueryParams struct {
	Type    string // consts.FeedTypeTimeline / FeedTypeFollowing / FeedTypeUser
	Content string // consts.FeedContentMoment / FeedContentArticle / FeedContentNovel / FeedContentAll
	Sort    string // consts.FeedSortTime / FeedSortHot（默认 time）
	UserID  uint32 // FeedTypeUser 时的目标用户 ID
	Offset  int
	Limit   int
}

// Result feed 组装结果。
type Result struct {
	Count int64
	Items []resp.FeedItem
}

// Build 组装 feed 列表。
// 登录态要求由 handler 层保证（following 未登录返回 40101）。
func Build(ctx context.Context, servants *server.Servants, uid uint32, params QueryParams) (*Result, error) {
	switch params.Type {
	case consts.FeedTypeTimeline:
		return buildTimeline(ctx, servants, uid, params)
	case consts.FeedTypeFollowing:
		return buildFollowing(ctx, servants, uid, params)
	case consts.FeedTypeUser:
		return buildUser(ctx, servants, uid, params)
	default:
		return nil, errors.New("unsupported feed type")
	}
}

// buildTimeline 时间线：全站公开内容按发布时间倒序。
func buildTimeline(ctx context.Context, servants *server.Servants, uid uint32, params QueryParams) (*Result, error) {
	moments, articles, err := queryBoth(ctx, servants, params, func(offset, limit int) ([]modelmoment.Moment, error) {
		return servants.ContentServant.QueryFeedMoments(offset, limit)
	}, func(offset, limit int) ([]modelarticle.Metadata, error) {
		return servants.ArticleServant.QueryFeedMetadatas(offset, limit)
	})
	if err != nil {
		return nil, err
	}

	items, err := assemble(ctx, servants, uid, moments, articles)
	if err != nil {
		return nil, err
	}

	count, err := countBoth(ctx, servants, params,
		servants.ContentServant.CountFeedMoments,
		servants.ArticleServant.CountFeedMetadatas)
	if err != nil {
		return nil, err
	}
	return &Result{Count: count, Items: items}, nil
}

// buildFollowing 关注流：仅关注用户的内容，按发布时间倒序。
// 动态可见性：公开全部可见；仅好友/仅粉丝需作者回关查看者；仅自己不进入关注流。
func buildFollowing(ctx context.Context, servants *server.Servants, uid uint32, params QueryParams) (*Result, error) {
	followingIDs, err := servants.InteractsServant.QueryFollowingIDs(uid)
	if err != nil {
		return nil, err
	}
	if len(followingIDs) == 0 {
		return &Result{Count: 0, Items: []resp.FeedItem{}}, nil
	}

	// 作者回关集合：用于"仅好友/仅粉丝"动态的可见性判定
	authorFollowsViewer, err := servants.InteractsServant.QueryFollowedByIDs(uid)
	if err != nil {
		return nil, err
	}

	moments, articles, err := queryBoth(ctx, servants, params, func(offset, limit int) ([]modelmoment.Moment, error) {
		return servants.ContentServant.QueryFeedMomentsByUsers(followingIDs, authorFollowsViewer, offset, limit)
	}, func(offset, limit int) ([]modelarticle.Metadata, error) {
		return servants.ArticleServant.QueryFeedMetadatasByUsers(followingIDs, offset, limit)
	})
	if err != nil {
		return nil, err
	}

	items, err := assemble(ctx, servants, uid, moments, articles)
	if err != nil {
		return nil, err
	}

	count, err := countBoth(ctx, servants, params,
		func() (int64, error) {
			return servants.ContentServant.CountFeedMomentsByUsers(followingIDs, authorFollowsViewer)
		},
		func() (int64, error) { return servants.ArticleServant.CountFeedMetadatasByUsers(followingIDs) })
	if err != nil {
		return nil, err
	}
	return &Result{Count: count, Items: items}, nil
}

// buildUser 用户内容流：某个用户发布的全部内容（动态 + 文章 + 小说元数据 + 章节数）。
// 动态按可见性（与关注流一致：公开全部可见，仅好友/仅粉丝需作者回关查看者，仅自己仅本人）；
// 文章/小说（元数据 + 章节数）按 status=normal 且非章节记录返回。
// 排序：time 按发布时间倒序；hot 按点赞量倒序（同点赞按发布时间倒序）。
// 拉黑/屏蔽/不想看校验由调用方（handler 中间件）完成。
func buildUser(ctx context.Context, servants *server.Servants, uid uint32, params QueryParams) (*Result, error) {
	targetID := params.UserID
	hot := params.Sort == consts.FeedSortHot

	// 动态可见性：isFriend（互相关注）/isFan（作者关注了查看者）
	var isFriend, isFan bool
	if uid != 0 && uid != targetID {
		following, err := servants.InteractsServant.IsFollowing(uid, targetID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		followedBy, err := servants.InteractsServant.IsFollowing(targetID, uid)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		isFriend, isFan = following && followedBy, followedBy
	}

	// hot 且 content=all 时：动态/文章各取 offset+limit 条，合并后全局热序再取 [offset, offset+limit)，
	// 避免"各自分页再拼接"导致第 N 页内容错位（单分类场景 SQL 层直接分页即可）。
	hotAll := hot && params.Content == consts.FeedContentAll
	fetchOffset, fetchLimit := params.Offset, params.Limit
	if hotAll {
		fetchLimit = params.Offset + params.Limit
	}

	var moments []modelmoment.Moment
	if params.Content == consts.FeedContentMoment || params.Content == consts.FeedContentAll {
		var err error
		if hot {
			moments, err = servants.ContentServant.QueryVisibleHotMomentsByUser(uid, targetID, isFriend, isFan, fetchOffset, fetchLimit)
		} else {
			moments, err = servants.ContentServant.QueryVisibleMomentsByUser(uid, targetID, isFriend, isFan, params.Offset, params.Limit)
		}
		if err != nil {
			return nil, err
		}
	}

	var articles []modelarticle.Metadata
	if params.Content == consts.FeedContentArticle || params.Content == consts.FeedContentNovel || params.Content == consts.FeedContentAll {
		var err error
		switch params.Content {
		case consts.FeedContentNovel:
			if hot {
				articles, err = servants.ArticleServant.QueryUserHotMetadatasByType(targetID, consts.ArticleTypeNovel, consts.ArticleStatusNormal, params.Offset, params.Limit)
			} else {
				articles, err = servants.ArticleServant.QueryUserNovels(targetID, consts.ArticleStatusNormal, params.Offset, params.Limit)
			}
		case consts.FeedContentArticle:
			if hot {
				articles, err = servants.ArticleServant.QueryUserHotMetadatasByType(targetID, consts.ArticleTypeNormal, consts.ArticleStatusNormal, params.Offset, params.Limit)
			} else {
				articles, err = servants.ArticleServant.QueryUserMetadatasByType(targetID, consts.ArticleTypeNormal, consts.ArticleStatusNormal, params.Offset, params.Limit)
			}
		default: // all
			if hot {
				articles, err = servants.ArticleServant.QueryUserHotMetadatas(targetID, consts.ArticleStatusNormal, fetchOffset, fetchLimit)
			} else {
				articles, err = servants.ArticleServant.QueryUserMetadatasAll(targetID, consts.ArticleStatusNormal, params.Offset, params.Limit)
			}
		}
		if err != nil {
			return nil, err
		}
	}

	items, err := assemble(ctx, servants, uid, moments, articles)
	if err != nil {
		return nil, err
	}

	// is_following：查看者是否关注了目标用户（用户内容流全部条目同属一个作者，查询一次回填）
	if uid != 0 && uid != targetID && len(items) > 0 {
		following, err := servants.InteractsServant.IsFollowing(uid, targetID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if following {
			for i := range items {
				items[i].IsFollowing = true
			}
		}
	}

	if hotAll {
		// 全局热序后截取本页
		sortHotItems(items)
		items = pageSlice(items, params.Offset, params.Limit)
	}

	// 小说章节数：批量统计后回填
	novelIDs := make([]uint64, 0, len(items))
	for i := range items {
		if items[i].Type == consts.ContentTypeArticle && items[i].ArticleType == consts.ArticleTypeNovel && items[i].NovelID == 0 {
			novelIDs = append(novelIDs, items[i].ID)
		}
	}
	if len(novelIDs) > 0 {
		chapterCounts, err := servants.ArticleServant.CountChaptersByNovels(novelIDs)
		if err != nil {
			return nil, err
		}
		for i := range items {
			if items[i].Type == consts.ContentTypeArticle {
				items[i].ChapterCount = chapterCounts[items[i].ID]
			}
		}
	}

	var count int64
	switch params.Content {
	case consts.FeedContentMoment:
		c, err := servants.ContentServant.CountVisibleMomentsByUser(uid, targetID, isFriend, isFan)
		if err != nil {
			return nil, err
		}
		count = c
	case consts.FeedContentArticle:
		c, err := servants.ArticleServant.CountUserMetadatasByType(targetID, consts.ArticleTypeNormal, consts.ArticleStatusNormal)
		if err != nil {
			return nil, err
		}
		count = c
	case consts.FeedContentNovel:
		c, err := servants.ArticleServant.CountUserNovels(targetID, consts.ArticleStatusNormal)
		if err != nil {
			return nil, err
		}
		count = c
	default: // all
		c1, err := servants.ContentServant.CountVisibleMomentsByUser(uid, targetID, isFriend, isFan)
		if err != nil {
			return nil, err
		}
		c2, err := servants.ArticleServant.CountUserMetadatasAll(targetID, consts.ArticleStatusNormal)
		if err != nil {
			return nil, err
		}
		count = c1 + c2
	}

	return &Result{Count: count, Items: items}, nil
}

// sortHotItems 按点赞量倒序排序 feed 条目（同点赞按发布时间倒序）。
func sortHotItems(items []resp.FeedItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Stats.Likes != items[j].Stats.Likes {
			return items[i].Stats.Likes > items[j].Stats.Likes
		}
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].ID > items[j].ID
	})
}

// pageSlice 返回已整体排序列表的第 offset 页（每页 limit 条）；越界返回空切片。
func pageSlice(items []resp.FeedItem, offset, limit int) []resp.FeedItem {
	if offset >= len(items) {
		return nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

// queryBoth 按 content 过滤参数分别查询动态与文章（各取 offset/limit 条，内存合并后截断）。
func queryBoth(ctx context.Context, servants *server.Servants, params QueryParams,
	momentQuery func(offset, limit int) ([]modelmoment.Moment, error),
	articleQuery func(offset, limit int) ([]modelarticle.Metadata, error),
) ([]modelmoment.Moment, []modelarticle.Metadata, error) {
	var (
		moments  []modelmoment.Moment
		articles []modelarticle.Metadata
	)
	switch params.Content {
	case consts.FeedContentMoment:
		m, err := momentQuery(params.Offset, params.Limit)
		if err != nil {
			return nil, nil, err
		}
		moments = m
	case consts.FeedContentArticle:
		a, err := articleQuery(params.Offset, params.Limit)
		if err != nil {
			return nil, nil, err
		}
		articles = a
	default: // FeedContentAll
		m, err := momentQuery(params.Offset, params.Limit)
		if err != nil {
			return nil, nil, err
		}
		a, err := articleQuery(params.Offset, params.Limit)
		if err != nil {
			return nil, nil, err
		}
		moments, articles = m, a
	}
	return moments, articles, nil
}

// countBoth 按 content 过滤参数统计动态与文章总数。
func countBoth(ctx context.Context, servants *server.Servants, params QueryParams,
	momentCount func() (int64, error),
	articleCount func() (int64, error),
) (int64, error) {
	switch params.Content {
	case consts.FeedContentMoment:
		return momentCount()
	case consts.FeedContentArticle:
		return articleCount()
	default:
		c1, err := momentCount()
		if err != nil {
			return 0, err
		}
		c2, err := articleCount()
		if err != nil {
			return 0, err
		}
		return c1 + c2, nil
	}
}

// assemble 组装 feed 条目：批量拉取作者/计数/点赞状态，过滤拉黑关系，按发布时间倒序合并。
func assemble(ctx context.Context, servants *server.Servants, uid uint32,
	moments []modelmoment.Moment, articles []modelarticle.Metadata,
) ([]resp.FeedItem, error) {
	// 1. 收集作者 ID 并批量查询
	authorIDs := make([]uint32, 0, len(moments)+len(articles))
	seen := map[uint32]bool{}
	collectAuthor := func(id uint32) {
		if id != 0 && !seen[id] {
			seen[id] = true
			authorIDs = append(authorIDs, id)
		}
	}
	for i := range moments {
		collectAuthor(moments[i].UserID)
	}
	for i := range articles {
		collectAuthor(articles[i].UserID)
	}

	authors, err := servants.UserServant.QueryUsersByIDs(authorIDs)
	if err != nil {
		return nil, err
	}

	// 2. 批量过滤拉黑/屏蔽/不想看关系（Redis pipeline 一次往返）
	filtered, err := servants.BlockServant.IsFilteredBatch(ctx, uid, authorIDs)
	if err != nil {
		return nil, err
	}

	// 3. 批量查询互动计数与点赞状态
	momentIDs := make([]uint64, 0, len(moments))
	for i := range moments {
		momentIDs = append(momentIDs, moments[i].ID)
	}
	articleIDs := make([]uint64, 0, len(articles))
	for i := range articles {
		articleIDs = append(articleIDs, articles[i].ID)
	}

	momentMetas, err := servants.ContentServant.QueryMomentInteractCounts(momentIDs)
	if err != nil {
		return nil, err
	}
	articleMetas, err := servants.ArticleServant.QueryInteractCounts(articleIDs)
	if err != nil {
		return nil, err
	}

	likedMoments := map[uint64]bool{}
	likedArticles := map[uint64]bool{}
	if uid != 0 {
		likedMoments, err = servants.InteractsServant.HasLikedMomentsBatch(uid, momentIDs)
		if err != nil {
			return nil, err
		}
		likedArticles, err = servants.InteractsServant.HasLikedArticlesBatch(uid, articleIDs)
		if err != nil {
			return nil, err
		}
	}

	// 4. 批量查询动态图片 UUID（feed 图片只下发 UUID，原始存储 URL 不下发）
	momentImages, err := servants.ContentServant.QueryMomentFileUUIDsBatch(momentIDs)
	if err != nil {
		return nil, err
	}

	// 5. 组装条目（跳过被过滤作者的内容）
	items := make([]resp.FeedItem, 0, len(moments)+len(articles))
	for i := range moments {
		m := &moments[i]
		if filtered[m.UserID] {
			continue
		}
		items = append(items, momentToItem(m, authors[m.UserID], momentMetas[m.ID], likedMoments[m.ID], momentImages[m.ID]))
	}
	for i := range articles {
		a := &articles[i]
		if filtered[a.UserID] {
			continue
		}
		items = append(items, articleToItem(a, authors[a.UserID], articleMetas[a.ID], likedArticles[a.ID]))
	}

	// 6. 按发布时间倒序合并（同时间按 id 倒序，保证稳定）
	sortFeedItems(items)
	return items, nil
}

// sortFeedItems 按发布时间倒序排序 feed 条目；同时间按 id 倒序（稳定）。
func sortFeedItems(items []resp.FeedItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].ID > items[j].ID
	})
}

// momentToItem 动态 → feed 条目。
// images 为动态图片文件 UUID 列表（不包含原始存储 URL，避免向客户端暴露存储细节）。
func momentToItem(m *modelmoment.Moment, author modeluser.User, meta modelmoment.MomentInteractCount, liked bool, images []string) resp.FeedItem {
	item := resp.FeedItem{
		ID:          m.ID,
		Type:        consts.ContentTypeMoment,
		UserID:      m.UserID,
		Title:       m.Title,
		Content:     m.Content,
		Images:      images,
		Status:      m.Status,
		Permission:  m.Permission,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Author:      author,
		IsLiked:     liked,
		IsFollowing: false,
		Full:        true,
	}
	item.Stats = resp.MomentStats{
		Likes:    meta.LikeCount,
		Comments: meta.CommentCount,
		Shares:   meta.ShareCount,
	}
	return item
}

// articleToItem 文章元数据 → feed 条目。
// 正文不触达 MongoDB（feed 只展示元数据与预览），Content 为空，Full=false。
func articleToItem(a *modelarticle.Metadata, author modeluser.User, meta modelarticle.ArticleInteractCount, liked bool) resp.FeedItem {
	item := resp.FeedItem{
		ID:          a.ID,
		Type:        consts.ContentTypeArticle,
		UserID:      a.UserID,
		Title:       a.Title,
		Description: a.Description,
		Cover:       a.Cover,
		ArticleType: a.Type,
		NovelID:     a.NovelID,
		ChapterID:   a.ChapterID,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		Author:      author,
		IsLiked:     liked,
		IsFollowing: false,
		Full:        false,
	}
	item.Stats = resp.MomentStats{
		Likes:    meta.LikeCount,
		Comments: meta.CommentCount,
		Shares:   meta.ShareCount,
	}
	return item
}
