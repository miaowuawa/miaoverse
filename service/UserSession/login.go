package UserSession

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"miaoverse/util/validate"
)

// loginUniversal 对于所有的登入会话都要使用这个函数处理
func loginUniversal(c fiber.Ctx) error {
	session.FromContext(c).Set("ClientType", validate.ParseUA(c.UserAgent()))
	session.FromContext(c).Set("UserAgent", c.UserAgent())
	return nil
}
