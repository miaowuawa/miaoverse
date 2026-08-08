package api

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/api/routers/v1/sms"
	userblock "miaoverse/api/routers/v1/user/block"
	usercomment "miaoverse/api/routers/v1/user/comment"
	usercontent "miaoverse/api/routers/v1/user/content"
	userfile "miaoverse/api/routers/v1/user/file"
	userinteract "miaoverse/api/routers/v1/user/interact"
	"miaoverse/api/routers/v1/user/login"
	usermoment "miaoverse/api/routers/v1/user/moment"
	"miaoverse/api/routers/v1/user/profile"
	userpunishment "miaoverse/api/routers/v1/user/punishment"
	userrelation "miaoverse/api/routers/v1/user/relation"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/server"
	"miaoverse/service/UserCheck"
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
	userGroup.Post("/moments", middleware.RequireNotPunished(servants, consts.PermPost), func(c fiber.Ctx) error {
		return usermoment.PublishHandler(c, servants)
	})
	// 评论动态。接口按内容类型（moments/...）划分，后续文章等类型使用各自的子路径。
	userGroup.Post("/moments/comments", middleware.RequireNotPunished(servants, consts.PermComment),
		middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.CreateHandler(c, servants)
		})
	// 回复动态下的评论（楼中楼）：先校验与动态作者的拉黑关系，再校验与被回复评论作者的拉黑关系。
	userGroup.Post("/moments/comments/:id/replies", middleware.RequireNotPunished(servants, consts.PermComment),
		middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentMomentAuthor, AllowSelf: true}),
		middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.ReplyHandler(c, servants)
		})
	// 获取楼中楼完整对话（传入楼中楼首条评论 id）。
	userGroup.Get("/moments/comments/:id/conversation",
		middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentMomentAuthor, AllowSelf: true}),
		middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.ConversationHandler(c, servants)
		})
	userGroup.Post("/follows", middleware.RequireNotPunished(servants, consts.PermSocial),
		middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{BodyField: "target", CheckMuteUnwatch: true}),
		func(c fiber.Ctx) error {
			return userinteract.FollowHandler(c, servants)
		})
	// 给动态点赞。接口按内容类型（moments/...）划分。
	userGroup.Post("/moments/likes", middleware.RequireNotPunished(servants, consts.PermSocial),
		middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return userinteract.LikeHandler(c, servants)
		})
	userGroup.Get("/punishments", func(c fiber.Ctx) error {
		return userpunishment.ListMineHandler(c, servants)
	})
	userGroup.Post("/blocks", func(c fiber.Ctx) error {
		return userblock.UpdateHandler(c, servants)
	})
	userGroup.Get("/users/:uid/contents/count", middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{PathParam: "uid"}),
		func(c fiber.Ctx) error {
			return usercontent.CountHandler(c, servants)
		})
	userGroup.Get("/users/:uid/contents", middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{PathParam: "uid"}),
		func(c fiber.Ctx) error {
			return usercontent.ListHandler(c, servants)
		})
	userGroup.Get("/users/:uid/info", func(c fiber.Ctx) error {
		return profile.GetUserInfoHandler(c, servants)
	})
	userGroup.Get("/users/:uid/following", middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{PathParam: "uid"}),
		func(c fiber.Ctx) error {
			return userrelation.FollowingHandler(c, servants)
		})
	userGroup.Get("/users/:uid/followers", middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{PathParam: "uid"}),
		func(c fiber.Ctx) error {
			return userrelation.FollowersHandler(c, servants)
		})
}
