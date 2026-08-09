package consts

// 评论域常量：目标类型、状态、内容长度限制、楼中楼对话深度。

const (
	CommentTargetMoment  uint8 = 1   // 动态
	CommentTargetArticle uint8 = 2   // 文章
	CommentTargetComment uint8 = 100 // 其他评论（楼中楼）
)

const (
	CommentStatusNormal     uint8 = 0 // 正常
	CommentStatusDeleted    uint8 = 1 // 删除
	CommentStatusDraft      uint8 = 2 // 草稿
	CommentStatusRestricted uint8 = 3 // 限制传播
	CommentStatusBlocked    uint8 = 4 // 屏蔽
)

const (
	MaxCommentLen        = 1000 // 评论/回复内容最大长度
	MaxConversationDepth = 10   // 楼中楼对话最大收集层数（一般 2-3 层即可覆盖）
	CommentChainMaxDepth = 20   // 楼中楼评论链上溯最大深度，防止脏数据导致死循环
)
