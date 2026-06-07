package middleware

import (
	"miaoverse/model/server"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/utils/v2"
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
}
