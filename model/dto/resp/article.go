package resp

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/service/i18n"
)

// ArticleStats 文章互动计数（面向详情页展示）。
type ArticleStats struct {
	Likes    uint32 `json:"likes"`
	Comments uint32 `json:"comments"`
	Shares   uint32 `json:"shares"`
	Views    uint32 `json:"views"`
}

// ArticleInfo 文章详情响应：元数据 + 正文 + 作者信息 + 互动计数 + 当前用户互动状态。
// 正文可能被截断（未登录前 60%/前 2 章），Full 表示返回的是否为完整正文。
type ArticleInfo struct {
	ID          uint64         `json:"id"`
	UserID      uint32         `json:"user_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Cover       string         `json:"cover"`
	Type        uint8          `json:"type"`
	NovelID     uint64         `json:"novel_id"`
	ChapterID   uint64         `json:"chapter_id"`
	Content     string         `json:"content"`
	Status      uint8          `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Author      modeluser.User `json:"author"`
	Stats       ArticleStats   `json:"stats"`
	IsLiked     bool           `json:"is_liked"`
	IsFollowing bool           `json:"is_following"`
	Full        bool           `json:"full"`
	Segments    int            `json:"segments"`
}

// ArticleSegment 文章正文分段响应：仅正文内容与分段序号。
type ArticleSegment struct {
	ID      uint64 `json:"id"`
	Seq     int    `json:"seq"`
	Content string `json:"content"`
	Full    bool   `json:"full"`
}

// CodeWithMsgArticle 文章详情响应体。
type CodeWithMsgArticle struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Article ArticleInfo `json:"article"`
}

// CodeWithMsgArticleSegment 文章分段响应体。
type CodeWithMsgArticleSegment struct {
	Code    int            `json:"code"`
	Msg     string         `json:"msg"`
	Segment ArticleSegment `json:"segment"`
}

// ArticleDetailOK 返回文章详情，作者已注销时展示字段统一打码。
func ArticleDetailOK(ctx fiber.Ctx, info ArticleInfo) error {
	MaskClosedAccount(ctx, &info.Author)
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgArticle{
		Code:    fiber.StatusOK,
		Msg:     i18n.Message(ctx, i18n.OKArticleFetched),
		Article: info,
	})
}

// ArticlePartialOK 返回截断后的文章详情（未登录前 60%/前 2 章），
// body 中 code 为自定义业务错误码 20001，提示登录后查看完整内容。
func ArticlePartialOK(ctx fiber.Ctx, info ArticleInfo) error {
	MaskClosedAccount(ctx, &info.Author)
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgArticle{
		Code:    consts.NeedLoginFullContent,
		Msg:     i18n.Message(ctx, i18n.OKArticleFetchedPartial),
		Article: info,
	})
}

// ArticleNeedSegmentsOK 返回文章详情但正文超长需分段拉取，
// body 中 code 为自定义业务错误码 20006，客户端应改用分段接口。
func ArticleNeedSegmentsOK(ctx fiber.Ctx, info ArticleInfo) error {
	MaskClosedAccount(ctx, &info.Author)
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgArticle{
		Code:    consts.ArticleNeedSegments,
		Msg:     i18n.Message(ctx, i18n.OKArticleNeedSegments),
		Article: info,
	})
}

// ArticleSegmentOK 返回文章正文分段。
func ArticleSegmentOK(ctx fiber.Ctx, segment ArticleSegment) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgArticleSegment{
		Code:    fiber.StatusOK,
		Msg:     i18n.Message(ctx, i18n.OKArticleSegmentFetched),
		Segment: segment,
	})
}
