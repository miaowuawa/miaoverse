package consts

// 自定义业务错误码。
// 规则：HTTP 状态码 + 两位序号，例如 40301 表示 HTTP 403 下的第 1 个业务错误。
// 所有被拉黑/拉黑对方导致的拒绝，body 中 code 统一为 BlockedByRelation。
const (
	BlockedByRelation = 40301 // 拉黑/被拉黑关系导致无法访问
	Punished          = 40302 // 权限封禁：账号封禁期间无法执行此操作
	AccountBanned     = 40303 // 账号封禁：账号被封禁，不允许登录/操作
	TargetPunished    = 40304 // 目标用户存在生效中的权限封禁
)

const (
	ContentBlocked = 45101 // 内容被屏蔽（动态/评论/文章等被标记为屏蔽状态），无法查看
)

// 2xx 成功响应下的业务码：HTTP 状态码仍为 200，客户端通过 body.code 识别细分语义。
const (
	NeedLoginFullContent = 20001 // 未登录只能查看文章前 60% / 小说前 2 章，登录后查看完整内容
	ArticleNeedSegments  = 20006 // 文章过长，需使用分段接口（/articles/:id/segments/:seq）拉取正文
)

// 4xx 下的业务码。
const (
	NeedLogin = 40101 // 需要登录：未登录访问需要登录的接口（如 following feed）
)
