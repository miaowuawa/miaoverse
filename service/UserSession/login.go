package UserSession

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

// loginUniversal 对于所有的登入会话都要使用这个函数处理
func loginUniversal(c fiber.Ctx) error {
	// 登录成功后重新生成 session ID，防止会话固定（session fixation）攻击：
	// 攻击者无法再使用登录前已知的 session ID 冒充用户。
	sess := session.FromContext(c)
	if sess == nil {
		return errors.New("session not found")
	}
	return sess.Regenerate()
}

// RefreshSessionID 在敏感操作（如修改密码）成功后重新生成 session ID，保留会话内容不变。
// 与登录一样用于防御会话固定：即使攻击者事先拿到了 session ID，操作后也会立即失效。
func RefreshSessionID(c fiber.Ctx) error {
	return loginUniversal(c)
}
