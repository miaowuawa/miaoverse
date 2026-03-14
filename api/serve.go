package api

import (
	"miaoverse/api/routers/v1/sms"
	"miaoverse/model/server"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Initial(app *fiber.App, servants *server.Servants) {
	// 健康检查
	app.Get("/", func(c fiber.Ctx) error {
		time := time.Now().Format("2006-01-02 15:04:05")
		return c.SendString("Miaoverse Content/User API Resp at " + time)
	})

	// API 路由组
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// 短信相关路由
	smsGroup := v1.Group("/sms")
	smsGroup.Post("/send", func(c fiber.Ctx) error {
		return sms.SendSmsHandler(c, servants.CodeManager, servants.SmsServant)
	})

}
