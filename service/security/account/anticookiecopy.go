package account

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"log"
	"miaoverse/model/dto/resp"
	"miaoverse/service/i18n"
)

func AntiCookieCopy(c fiber.Ctx) error {
	// 1. 获取 Session 实例（先判空，避免 nil 调用）
	sess := session.FromContext(c)
	if sess == nil {
		return c.Next()
	}
	if sess.Get("UID") == nil {
		return c.Next()
	}
	// 手机版是特殊情况，每次改version会变ua，这个函数中不处理
	if sess.Get("ClientType") == "mobile" {
		return c.Next()
	}

	// 2. 获取存储的 UA（未存储时 ua 为 nil）
	ua := sess.Get("UserAgent")

	// 3. 核心逻辑：
	// - UA 未存储 → 拒绝（登录时未记录 UA，异常）
	// - UA 存储但与当前请求不一致 → 拒绝（环境变化）
	uaStr, ok := ua.(string)
	if !ok || c.UserAgent() != uaStr {
		// ① 输出错误日志（TODO：替换为你的日志组件）
		log.Printf("AntiCookieCopy 触发：存储UA=%v，当前UA=%s", ua, c.UserAgent())

		// ② 销毁 Session（不再忽略错误）
		if err := sess.Destroy(); err != nil {
			log.Printf("Session 销毁失败：%v", err)
		}

		// ③ 返回 403 响应并终止请求（关键：return 而非 _）
		return c.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  i18n.Message(c, i18n.ErrLoginEnvironmentChange),
		})
	}

	// 4. 验证通过 → 执行后续中间件/处理器
	return c.Next()
}
