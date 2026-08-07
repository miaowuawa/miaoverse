package api

import (
	"miaoverse/api/routers/v1/sms"
	userfile "miaoverse/api/routers/v1/user/file"
	"miaoverse/api/routers/v1/user/login"
	usermoment "miaoverse/api/routers/v1/user/moment"
	"miaoverse/api/routers/v1/user/profile"
	"miaoverse/middleware"
	"miaoverse/model/server"
	"miaoverse/service/UserCheck"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Initial(app *fiber.App, servants *server.Servants) {
	middleware.Initial(app, servants)

	// 健康检查
	app.Get("/", func(c fiber.Ctx) error {
		time := time.Now().Format("2006-01-02 15:04:05")
		return c.SendString("Miaoverse Content/User API Resp at " + time)
	})

	// API 路由组
	api := app.Group("/api")
	// v1路由组
	v1 := api.Group("/v1")

	// 短信相关路由
	smsGroup := v1.Group("/sms")
	smsGroup.Post("/send", func(c fiber.Ctx) error {
		return sms.SendSmsHandler(c, servants)
	})
	loginGroup := v1.Group("/user/login")
	loginGroup.Post("/sms", func(c fiber.Ctx) error {
		return login.BySMSHandler(c, servants)
	})
	loginGroup.Post("/choose", func(c fiber.Ctx) error {
		return login.ChooseUserHandler(c, servants)
	})

	registerGroup := v1.Group("/user/register")
	registerGroup.Post("/sms", func(c fiber.Ctx) error {
		return login.RegisterBySMSHandler(c, servants)
	})

	userGroup := v1.Group("/user")
	userGroup.Use(middleware.RequireUser(servants, UserCheck.AccountActive()))
	userGroup.Put("/info", func(c fiber.Ctx) error {
		return profile.UpdateFullHandler(c, servants)
	})
	userGroup.Patch("/info", func(c fiber.Ctx) error {
		return profile.UpdatePartialHandler(c, servants)
	})
	userGroup.Post("/files", func(c fiber.Ctx) error {
		return userfile.UploadHandler(c, servants)
	})
	userGroup.Get("/files/:uuid/temp-link", func(c fiber.Ctx) error {
		return userfile.TempLinkHandler(c, servants)
	})
	userGroup.Get("/files/:uuid/shared-link", func(c fiber.Ctx) error {
		return userfile.SharedTempLinkHandler(c, servants)
	})
	userGroup.Post("/moments", func(c fiber.Ctx) error {
		return usermoment.PublishHandler(c, servants)
	})
}
