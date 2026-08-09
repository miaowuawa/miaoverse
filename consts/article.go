package consts

// 文章（article）域常量：状态、类型、长度限制、MongoDB 集合名。

const (
	ArticleStatusNormal     uint8 = 0 // 正常
	ArticleStatusDeleted    uint8 = 1 // 删除
	ArticleStatusDraft      uint8 = 2 // 草稿
	ArticleStatusRestricted uint8 = 3 // 限制传播
	ArticleStatusBlocked    uint8 = 4 // 屏蔽
)

const (
	ArticleTypeNormal uint8 = 0 // 普通文章
	ArticleTypeRepost uint8 = 1 // 转载/分享的文章
	ArticleTypeNovel  uint8 = 2 // 小说（分章节，chapter_id 生效）
)

const (
	MaxArticleTitleLen       = 255   // 标题最大长度（字符）
	MaxArticleDescriptionLen = 500   // 摘要最大长度（字符）
	MaxArticlePreviewHeadLen = 300   // 预览开头最大长度（字符）
	MaxArticleCoverLen       = 255   // 封面 URL 最大长度
	MaxArticleContentLen     = 50000 // 正文最大长度（字符）
	MaxArticleReferenceLen   = 100   // 引用文章最大数量
	MaxArticleChapterID      = 10000 // 章节号上限（防止异常章节号）
)

// ArticleMongoCollection MongoDB 中文章正文档集合名
const ArticleMongoCollection = "article"

// ContentTypeArticle 内容类型标识（供互动/评论域识别）
const ContentTypeArticle = "article"
