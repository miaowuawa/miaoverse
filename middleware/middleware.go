package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/utils/v2"
	"miaoverse/model/server"
	"miaoverse/service/security/account"
	"time"
)

func Initial(app *fiber.App, servant *server.Servants) {
	//requestID

	app.Use(requestid.New(requestid.Config{
		Header:    "X-Miaoverse-ReqID",
		Generator: utils.SecureToken,
	}))

	//account
	app.Use(session.New(
		session.Config{
			CookieHTTPOnly:  true,                 // Prevent XSS
			CookieSameSite:  "Lax",                // CSRF protection
			IdleTimeout:     30 * 24 * time.Hour,  // Session timeout
			AbsoluteTimeout: 120 * 24 * time.Hour, // Maximum account life
			Extractor:       extractors.FromCookie("mwu_sess_id"),
			Storage:         servant.FiberSessionStorage,
		}))

	app.Use(func(ctx fiber.Ctx) error {
		return account.AntiCookieCopy(ctx)
	})
}
