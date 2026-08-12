package resp

import (
	"time"

	modeluser "miaoverse/model/dao/user"
)

// FeedItem feed 条目：动态与文章统一结构。
// 动态：Content 为动态正文，无 Description/Cover/Type/NovelID/ChapterID（零值）；
// Images 为动态图片文件 UUID 列表（仅动态有值，文章为空），原始存储 URL 不下发，需经临时链接接口换取。
// 文章：Content 为文章正文（可能被截断，Full 表示是否完整），Description/Cover 等为元数据字段。
type FeedItem struct {
	ID          uint64         `json:"id"`
	Type        string         `json:"type"` // moment / article
	UserID      uint32         `json:"user_id"`
	Title       string         `json:"title"`
	Content     string         `json:"content"`
	Description string         `json:"description"`
	Cover       string         `json:"cover"`
	Images      []string       `json:"images"`
	ArticleType uint8          `json:"article_type"`
	NovelID     uint64         `json:"novel_id"`
	ChapterID   uint64         `json:"chapter_id"`
	Status      uint8          `json:"status"`
	Permission  uint8          `json:"permission"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Author      modeluser.User `json:"author"`
	Stats       MomentStats    `json:"stats"`
	IsLiked     bool           `json:"is_liked"`
	IsFollowing bool           `json:"is_following"`
	Full        bool           `json:"full"`
}

// CodeWithMsgFeedList feed 列表响应体。
type CodeWithMsgFeedList struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg"`
	Count int64      `json:"count"`
	Items []FeedItem `json:"items"`
}
