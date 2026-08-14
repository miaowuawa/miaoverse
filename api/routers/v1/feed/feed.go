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
// 路由：GET /api/v1/feeds/:type?content=&sort=&offset=&limit=
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

	params, ok := parseParams(ctx, feedType)
	if !ok {
		return resp.BadRequest(ctx)
	}

	result, err := Feed.Build(ctx.Context(), servants, uid, params)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.FeedList(ctx, result.Count, result.Items)
}

// UserListHandler 用户内容流：GET /api/v1/feeds/user/:uid?content=&sort=&offset=&limit=
// 目标用户 ID 由路由中间件 RequireNoBlockUser(PathParam: "uid", AllowSelf: true) 解析并缓存，
// 这里读取解析结果，避免重复解析与重复校验；账号封禁校验在 handler 内完成（40303）。
func UserListHandler(ctx fiber.Ctx, servants *server.Servants) error {
	targetID, ok := middleware.BlockTarget(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	// 目标用户处于账号封禁状态（不允许登录）时返回 40303，禁止查看其内容
	user, err := servants.UserServant.QueryByID(targetID)
	if err != nil {
		return resp.UserNotFound(ctx)
	}
	if user.Status == consts.UserStatusBanned {
		return resp.AccountBanned(ctx)
	}

	params, ok := parseParams(ctx, consts.FeedTypeUser)
	if !ok {
		return resp.BadRequest(ctx)
	}
	params.UserID = targetID

	uid, _ := middleware.CurrentUID(ctx)
	result, err := Feed.Build(ctx.Context(), servants, uid, params)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.FeedList(ctx, result.Count, result.Items)
}

// parseParams 解析 feed 公共查询参数（content/sort/offset/limit）。
// novel 内容类型与 hot 排序仅用户内容流支持（timeline/following 返回 false 走 400）。
func parseParams(ctx fiber.Ctx, feedType string) (Feed.QueryParams, bool) {
	content := strings.TrimSpace(ctx.Query("content"))
	if content == "" {
		content = consts.FeedContentAll
	}
	if content != consts.FeedContentMoment && content != consts.FeedContentArticle &&
		content != consts.FeedContentNovel && content != consts.FeedContentAll {
		return Feed.QueryParams{}, false
	}
	if feedType != consts.FeedTypeUser && content == consts.FeedContentNovel {
		return Feed.QueryParams{}, false
	}

	sortBy := strings.TrimSpace(ctx.Query("sort"))
	if sortBy == "" {
		sortBy = consts.FeedSortTime
	}
	if sortBy != consts.FeedSortTime && sortBy != consts.FeedSortHot {
		return Feed.QueryParams{}, false
	}
	if feedType != consts.FeedTypeUser && sortBy == consts.FeedSortHot {
		return Feed.QueryParams{}, false
	}

	offset, limit, ok := pagination.Parse(ctx.Query("offset"), ctx.Query("limit"))
	if !ok {
		return Feed.QueryParams{}, false
	}
	if limit > consts.FeedMaxLimit {
		limit = consts.FeedMaxLimit
	}

	return Feed.QueryParams{
		Type:    feedType,
		Content: content,
		Sort:    sortBy,
		Offset:  offset,
		Limit:   limit,
	}, true
}
