package article

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/middleware"
	modelarticle "miaoverse/model/dao/article"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/Article"
)

// DetailHandler 获取文章详情（含正文）。内容屏蔽校验由 RequireNoContentBlock（45101）、
// 拉黑/被拉黑校验由 RequireNoBlockUser（40301）中间件完成，元数据由 ResolveArticlePathAuthor 解析并缓存。
// 未登录仅可查看正文前 60%（小说前 2 章），body code 20001；正文超长时 body code 20006 引导分段拉取。
func DetailHandler(ctx fiber.Ctx, servants *server.Servants) error {
	meta, ok := middleware.BlockArticleMeta(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
	}

	uid, loggedIn := middleware.CurrentUID(ctx)

	result, err := Article.BuildDetail(ctx.Context(), servants, meta, loggedIn)
	if err != nil {
		return resp.ServerError(ctx)
	}

	info, err := composeDetail(ctx, servants, meta, result, uid, loggedIn)
	if err != nil {
		return err
	}

	switch {
	case result.Segments > 1:
		return resp.ArticleNeedSegmentsOK(ctx, info)
	case !result.Full:
		return resp.ArticlePartialOK(ctx, info)
	default:
		return resp.ArticleDetailOK(ctx, info)
	}
}

// SegmentHandler 分段获取文章正文（seq 从 1 起），与详情接口共用屏蔽/拉黑校验与截断口径。
// 未登录仅可拉取正文前 60%（小说前 2 章）范围内的分段。
func SegmentHandler(ctx fiber.Ctx, servants *server.Servants) error {
	meta, ok := middleware.BlockArticleMeta(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
	}

	seq, err := strconv.Atoi(strings.TrimSpace(ctx.Params("seq")))
	if err != nil {
		return resp.BadRequest(ctx)
	}

	_, loggedIn := middleware.CurrentUID(ctx)
	segment, ok, err := Article.BuildSegment(ctx.Context(), servants, meta, seq, loggedIn)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if !ok {
		return resp.BadRequest(ctx)
	}

	return resp.ArticleSegmentOK(ctx, segment)
}

// composeDetail 组装文章详情响应：作者信息 + 当前用户互动状态。
// 作者查询失败时按 404 处理（作者已注销/不存在），与动态详情口径一致。
func composeDetail(ctx fiber.Ctx, servants *server.Servants, meta *modelarticle.Metadata, result *Article.DetailResult, uid uint32, loggedIn bool) (resp.ArticleInfo, error) {
	author, err := servants.UserServant.QueryByID(meta.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ArticleInfo{}, resp.FileNotFound(ctx)
		}
		return resp.ArticleInfo{}, resp.ServerError(ctx)
	}

	isLiked, isFollowing := false, false
	if loggedIn {
		isLiked, err = servants.InteractsServant.HasLikedMoment(uid, meta.ID)
		if err != nil {
			return resp.ArticleInfo{}, resp.ServerError(ctx)
		}
		isFollowing, err = servants.InteractsServant.IsFollowing(uid, meta.UserID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ArticleInfo{}, resp.ServerError(ctx)
		}
	}

	return Article.ToArticleInfo(meta, result, author, isLiked, isFollowing), nil
}
