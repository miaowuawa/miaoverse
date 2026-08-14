package consts

// Feed 域常量：feed 类型、内容类型过滤、排序、分页限制。

const (
	FeedTypeTimeline  = "timeline"  // 时间线：全站公开内容按发布时间倒序
	FeedTypeFollowing = "following" // 关注流：按关注用户的最新内容排序
	FeedTypeUser      = "user"      // 用户内容流：某个用户发布的全部内容
)

const (
	FeedContentMoment  = "moment"  // 只拉取动态
	FeedContentArticle = "article" // 只拉取文章（含小说根文章与普通文章）
	FeedContentNovel   = "novel"   // 只拉取小说（type=novel 的文章）
	FeedContentAll     = "all"     // 全部
)

const (
	FeedSortTime = "time" // 按发布时间倒序
	FeedSortHot  = "hot"  // 按点赞量倒序（同点赞按时间倒序）
)

const (
	FeedMaxLimit = 50 // feed 单页最大条数（feed 条目含作者信息，比普通列表更重）
)
