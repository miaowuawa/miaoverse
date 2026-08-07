package api

import (
	"miaoverse/api/routers/v1/sms"
	userblock "miaoverse/api/routers/v1/user/block"
	usercomment "miaoverse/api/routers/v1/user/comment"
	usercontent "miaoverse/api/routers/v1/user/content"
	userfile "miaoverse/api/routers/v1/user/file"
	"miaoverse/api/routers/v1/user/login"
	usermoment "miaoverse/api/routers/v1/user/moment"
	"miaoverse/api/routers/v1/user/profile"
	userpunishment "miaoverse/api/routers/v1/user/punishment"
	userrelation "miaoverse/api/routers/v1/user/relation"
	"miaoverse/middleware"
	modeluser "miaoverse/model/dao/user"
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
	userGroup.Post("/moments", middleware.RequireNotPunished(servants, modeluser.PermPost), func(c fiber.Ctx) error {
		return usermoment.PublishHandler(c, servants)
	})
	userGroup.Post("/comments", middleware.RequireNotPunished(servants, modeluser.PermComment), func(c fiber.Ctx) error {
		return usercomment.CreateHandler(c, servants)
	})
	userGroup.Get("/punishments", func(c fiber.Ctx) error {
		return userpunishment.ListMineHandler(c, servants)
	})
	userGroup.Post("/blocks", func(c fiber.Ctx) error {
		return userblock.UpdateHandler(c, servants)
	})
	userGroup.Get("/users/:uid/contents/count", func(c fiber.Ctx) error {
		return usercontent.CountHandler(c, servants)
	})
	userGroup.Get("/users/:uid/contents", func(c fiber.Ctx) error {
		return usercontent.ListHandler(c, servants)
	})
	userGroup.Get("/users/:uid/info", func(c fiber.Ctx) error {
		return profile.GetUserInfoHandler(c, servants)
	})
	userGroup.Get("/users/:uid/following", func(c fiber.Ctx) error {
		return userrelation.FollowingHandler(c, servants)
	})
	userGroup.Get("/users/:uid/followers", func(c fiber.Ctx) error {
		return userrelation.FollowersHandler(c, servants)
	})
}
