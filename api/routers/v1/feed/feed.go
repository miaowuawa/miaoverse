package feed

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/Feed"
	"miaoverse/util/pagination"
)

// ListHandler 拉取 feed 列表。
// 路由：GET /api/v1/feeds/:type?content=&offset=&limit=
//   - type=timeline：时间线，无需登录；
//   - type=following：关注流，未登录返回 40101。
func ListHandler(ctx fiber.Ctx, servants *server.Servants) error {
	feedType := strings.TrimSpace(ctx.Params("type"))
	if feedType != consts.FeedTypeTimeline && feedType != consts.FeedTypeFollowing {
		return resp.BadRequest(ctx)
	}

	uid, loggedIn := middleware.CurrentUID(ctx)
	if feedType == consts.FeedTypeFollowing && !loggedIn {
		return resp.NeedLogin(ctx)
	}

	content := strings.TrimSpace(ctx.Query("content"))
	if content == "" {
		content = consts.FeedContentAll
	}
	if content != consts.FeedContentMoment && content != consts.FeedContentArticle && content != consts.FeedContentAll {
		return resp.BadRequest(ctx)
	}

	offset, limit, ok := pagination.Parse(ctx.Query("offset"), ctx.Query("limit"))
	if !ok {
		return resp.BadRequest(ctx)
	}
	if limit > consts.FeedMaxLimit {
		limit = consts.FeedMaxLimit
	}

	result, err := Feed.Build(ctx.Context(), servants, uid, Feed.QueryParams{
		Type:    feedType,
		Content: content,
		Offset:  offset,
		Limit:   limit,
	})
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.FeedList(ctx, result.Count, result.Items)
}
