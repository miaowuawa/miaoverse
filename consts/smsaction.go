package consts

// 短信验证码业务场景（action）。
// 验证码在 Redis 中按「场景 + 区号 + 手机号」隔离存储：
// 为登录申请的验证码不能用于修改密码，反之亦然，避免验证码被跨场景重放。
const (
	ActionLogin          = "LOGIN"           // 登录 / 注册（公开接口即可申请）
	ActionChangePassword = "CHANGE_PASSWORD" // 修改密码（需登录，手机号取自会话）
)

// PublicSMSActions 无需登录即可申请验证码的业务场景白名单。
//
// 刻意不包含 ActionChangePassword：修改密码验证码只能通过登录态接口
// POST /api/v1/user/password/sms 申请，手机号由服务端从会话取。
// 这样攻击者无法用公开接口给任意手机号发送「修改密码」短信（短信轰炸 / 社工诱导）。
var PublicSMSActions = []string{ActionLogin}

// IsPublicSMSAction 判断 action 是否属于公开可申请的场景。
func IsPublicSMSAction(action string) bool {
	for _, allowed := range PublicSMSActions {
		if allowed == action {
			return true
		}
	}
	return false
}
