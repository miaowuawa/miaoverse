package consts

// Feed 域常量：feed 类型、内容类型过滤、分页限制。

const (
	FeedTypeTimeline  = "timeline"  // 时间线：按发布时间倒序
	FeedTypeFollowing = "following" // 关注流：按关注用户的最新内容排序
)

const (
	FeedContentMoment  = "moment"  // 只拉取动态
	FeedContentArticle = "article" // 只拉取文章
	FeedContentAll     = "all"     // 动态 + 文章
)

const (
	FeedMaxLimit = 50 // feed 单页最大条数（feed 条目含作者信息，比普通列表更重）
)
