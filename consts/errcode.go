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
