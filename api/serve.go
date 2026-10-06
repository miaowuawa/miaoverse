package api

import (
	userarticle "miaoverse/api/routers/v1/article"
	usercontent "miaoverse/api/routers/v1/content"
	userfeed "miaoverse/api/routers/v1/feed"
	usermoment "miaoverse/api/routers/v1/moment"
	usercomment "miaoverse/api/routers/v1/moment/comment"
	usersticker "miaoverse/api/routers/v1/sticker"
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/api/routers/v1/sms"
	userblock "miaoverse/api/routers/v1/user/block"
	userfile "miaoverse/api/routers/v1/user/file"
	userinteract "miaoverse/api/routers/v1/user/interact"
	"miaoverse/api/routers/v1/user/login"
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
		timeNow := time.Now().Format("2006-01-02 15:04:05")
		return c.SendString("Miaoverse Content/User API Resp at " + timeNow)
	})

	// API 路由组
	api := app.Group("/api")
	// v1路由组
	v1 := api.Group("/v1")

	// ===== 认证组 /auth（无需登录）=====
	authGroup := v1.Group("/auth")
	authGroup.Post("/sms/send", func(c fiber.Ctx) error {
		return sms.SendSmsHandler(c, servants)
	})
	authGroup.Post("/login/sms", func(c fiber.Ctx) error {
		return login.BySMSHandler(c, servants)
	})
	authGroup.Post("/login/choose", func(c fiber.Ctx) error {
		return login.ChooseUserHandler(c, servants)
	})
	// 账号切换：列表与切换接口不校验当前账号状态（账号被封禁也可以切换到同手机号下的其他正常账号）。
	authGroup.Get("/accounts", func(c fiber.Ctx) error {
		return login.AccountListHandler(c, servants)
	})
	authGroup.Post("/switch", func(c fiber.Ctx) error {
		return login.SwitchAccountHandler(c, servants)
	})
	authGroup.Post("/register/sms", func(c fiber.Ctx) error {
		return login.RegisterBySMSHandler(c, servants)
	})
	authGroup.Post("/logout", func(c fiber.Ctx) error {
		return login.LogoutHandler(c, servants)
	})

	// ===== Feed 组 /feeds =====
	// timeline 时间线（无需登录）：全站公开内容按发布时间倒序。
	// following 关注流（需登录，未登录返回 40101）：仅关注用户的内容。
	// user 用户内容流（需登录）：?uid= 指定用户发布的全部内容（动态/文章/小说元数据+章节数），
	//     按 ?sort=time|hot 排序；拉黑/被拉黑/屏蔽/不想看关系与账号封禁校验在此路由完成（40301/40303）。
	// content 过滤：moment 只拉动态 / article 只拉文章 / novel 只拉小说 / all 动态+文章（默认）。
	v1.Get("/feeds/:type", func(c fiber.Ctx) error {
		return userfeed.ListHandler(c, servants)
	})
	// 用户内容流单独路由：挂拉黑校验中间件（PathParam: uid），并允许本人查看。
	// 目标用户账号封禁（UserStatusBanned）由 handler 内校验返回 40303。
	v1.Get("/feeds/user/:uid", middleware.RequireUser(servants, UserCheck.AccountActive()),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{PathParam: "uid", AllowSelf: true}),
		func(c fiber.Ctx) error {
			return userfeed.UserListHandler(c, servants)
		})

	// ===== 动态组 /moment（需登录）=====
	momentGroup := v1.Group("/moment")
	momentGroup.Use(middleware.RequireUser(servants, UserCheck.AccountActive()))
	// 发布动态
	momentGroup.Post("/", middleware.RequireNotPunished(servants, consts.PermPost), func(c fiber.Ctx) error {
		return usermoment.PublishHandler(c, servants)
	})
	// 编辑动态（仅作者本人）：先校验内容是否被屏蔽，再校验与作者的拉黑关系（AllowSelf 允许本人）。
	// 账号封禁与发布权限封禁校验与发布接口一致。
	momentGroup.Patch("/:id", middleware.RequireNotPunished(servants, consts.PermPost),
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentPathAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentPathAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usermoment.UpdateHandler(c, servants)
		})
	// 给动态点赞。接口按内容类型（/moment/...）划分，后续文章等类型使用各自的子路径。
	momentGroup.Post("/likes", middleware.RequireNotPunished(servants, consts.PermSocial),
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return userinteract.LikeHandler(c, servants)
		})
	// 取消点赞（幂等）。校验与点赞一致。
	momentGroup.Delete("/likes", middleware.RequireNotPunished(servants, consts.PermSocial),
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return userinteract.UnlikeHandler(c, servants)
		})

	// 动态详情（无需登录）：未登录用户仅可查看公开动态；登录用户按可见性规则查看。
	// 内容被屏蔽返回 451（45101）；拉黑/被拉黑任意一方存在时拒绝（40301）；动态不存在或不可见返回 404。
	v1.Get("/moments/:id",
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentPathAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{
			Resolver:       middleware.ResolveMomentPathAuthor,
			AllowSelf:      true,
			AllowAnonymous: true,
		}), func(c fiber.Ctx) error {
			return usermoment.DetailHandler(c, servants)
		})

	// 文章详情（无需登录）：未登录仅可查看正文前 60%（小说前 2 章），body code 20001；正文超长返回 20006 引导分段。
	// 内容被屏蔽返回 451（45101）；拉黑/被拉黑任意一方存在时拒绝（40301）；文章不存在或非正常状态返回 404。
	v1.Get("/articles/:id",
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveArticlePathAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{
			Resolver:       middleware.ResolveArticlePathAuthor,
			AllowSelf:      true,
			AllowAnonymous: true,
		}), func(c fiber.Ctx) error {
			return userarticle.DetailHandler(c, servants)
		})
	// 文章正文分段（无需登录）：seq 从 1 起；未登录仅可拉取正文前 60%（小说前 2 章）范围内的分段。
	v1.Get("/articles/:id/segments/:seq",
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveArticlePathAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{
			Resolver:       middleware.ResolveArticlePathAuthor,
			AllowSelf:      true,
			AllowAnonymous: true,
		}), func(c fiber.Ctx) error {
			return userarticle.SegmentHandler(c, servants)
		})

	// ===== 评论组 /comment（需登录）=====
	// 接口按被评论的内容类型划分子路径（/comment/moments、后续 /comment/articles），
	// 不提供通用 /comment 根接口，避免不同类型评论混淆。
	commentGroup := v1.Group("/comment")
	commentGroup.Use(middleware.RequireUser(servants, UserCheck.AccountActive()))
	// 评论动态
	commentGroup.Post("/moments", middleware.RequireNotPunished(servants, consts.PermComment),
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.CreateHandler(c, servants)
		})
	// 回复动态下的评论（楼中楼）：先校验内容是否被屏蔽，再校验与动态作者的拉黑关系，再校验与被回复评论作者的拉黑关系。
	commentGroup.Post("/moments/:id/replies", middleware.RequireNotPunished(servants, consts.PermComment),
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentMomentAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentMomentAuthor, AllowSelf: true}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.ReplyHandler(c, servants)
		})
	// 给评论点赞。校验与回复评论一致：先校验内容是否被屏蔽（评论及其上溯链/所属动态），再校验与评论作者的拉黑关系。
	commentGroup.Post("/likes", middleware.RequireNotPunished(servants, consts.PermSocial),
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentBodyAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentBodyAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.LikeCommentHandler(c, servants)
		})
	// 取消评论点赞（幂等）。校验与点赞一致。
	commentGroup.Delete("/likes", middleware.RequireNotPunished(servants, consts.PermSocial),
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentBodyAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentBodyAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.UnlikeCommentHandler(c, servants)
		})
	// 获取楼中楼完整对话（传入楼中楼首条评论 id）。
	commentGroup.Get("/moments/:id/conversation",
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentMomentAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentMomentAuthor, AllowSelf: true}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveCommentAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.ConversationHandler(c, servants)
		})
	// 获取动态一级评论分页列表（:id 为动态 id；sort=hot|time）。
	// 屏蔽/拉黑校验与动态详情一致；评论可内嵌贴纸（[sticker:<uuid>] 标记随文本穿插展示）。
	commentGroup.Get("/moments/:id",
		middleware.RequireNoContentBlock(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentPathAuthor}),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{Resolver: middleware.ResolveMomentPathAuthor, AllowSelf: true}),
		func(c fiber.Ctx) error {
			return usercomment.ListHandler(c, servants)
		})

	// ===== 贴纸组 /stickers（需登录 + 已绑定手机号）=====
	// 分为「贴纸收藏夹（个人）」与「贴纸包（其他贴纸包）」两部分：
	// 收藏夹存放本人上传（source=1，上传时自动加入）与收藏他人（source=2）的贴纸，可置顶，收藏上限 500；
	// 贴纸包由具备评论权限的用户创建，单包上限 100 张，可整包收藏（收藏后包内容更新自动同步）。
	// 贴纸包被封禁（sticker_pack.banned，运营处置）后，包内贴纸在所有使用处无法显示。
	stickerGroup := v1.Group("/stickers")
	stickerGroup.Use(middleware.RequireUser(servants, UserCheck.AccountActive(), UserCheck.PhoneBound()))
	// 上传贴纸（multipart/form-data：file + 可选 name），单张大小从配置读取（默认 10MB）
	stickerGroup.Post("/", middleware.RequireNotPunished(servants, consts.PermUploadFile), func(c fiber.Ctx) error {
		return usersticker.UploadHandler(c, servants)
	})
	// 我的贴纸收藏夹列表（置顶优先）
	stickerGroup.Get("/collection", func(c fiber.Ctx) error {
		return usersticker.CollectionHandler(c, servants)
	})
	// 把他人贴纸添加到收藏夹（幂等）
	stickerGroup.Post("/collection", func(c fiber.Ctx) error {
		return usersticker.FavoriteHandler(c, servants)
	})
	// 从收藏夹移除收藏的贴纸（幂等）
	stickerGroup.Delete("/collection", func(c fiber.Ctx) error {
		return usersticker.UnfavoriteHandler(c, servants)
	})
	// 设置/取消收藏夹贴纸置顶（body: top 0|1）
	stickerGroup.Patch("/:uuid/top", func(c fiber.Ctx) error {
		return usersticker.TopHandler(c, servants)
	})
	// 删除本人上传的贴纸
	stickerGroup.Delete("/:uuid", func(c fiber.Ctx) error {
		return usersticker.DeleteHandler(c, servants)
	})
	// 贴纸包列表（filter=all 全部 / favorite 已整包收藏）
	stickerGroup.Get("/packs", func(c fiber.Ctx) error {
		return usersticker.PackListHandler(c, servants)
	})
	// 创建贴纸包（要求未被封禁评论权限）
	stickerGroup.Post("/packs", middleware.RequireNotPunished(servants, consts.PermComment), func(c fiber.Ctx) error {
		return usersticker.PackCreateHandler(c, servants)
	})
	// 收藏/取消收藏整个贴纸包（幂等）
	stickerGroup.Post("/packs/favorites", func(c fiber.Ctx) error {
		return usersticker.PackFavoriteHandler(c, servants)
	})
	stickerGroup.Delete("/packs/favorites", func(c fiber.Ctx) error {
		return usersticker.PackUnfavoriteHandler(c, servants)
	})
	// 贴纸包详情（含包内贴纸列表）
	stickerGroup.Get("/packs/:id", func(c fiber.Ctx) error {
		return usersticker.PackDetailHandler(c, servants)
	})
	// 向自己的贴纸包添加/移出贴纸
	stickerGroup.Post("/packs/:id/stickers", func(c fiber.Ctx) error {
		return usersticker.PackAddStickerHandler(c, servants)
	})
	stickerGroup.Delete("/packs/:id/stickers/:sticker_uuid", func(c fiber.Ctx) error {
		return usersticker.PackRemoveStickerHandler(c, servants)
	})

	// ===== 用户组 /user（需登录）=====
	userGroup := v1.Group("/user")
	userGroup.Use(middleware.RequireUser(servants, UserCheck.AccountActive()))
	userGroup.Put("/info", func(c fiber.Ctx) error {
		return profile.UpdateFullHandler(c, servants)
	})
	userGroup.Patch("/info", func(c fiber.Ctx) error {
		return profile.UpdatePartialHandler(c, servants)
	})
	// 设置头像：头像文件必须是本人 active 图片且公开（permission=0），
	// 修改头像需要未被封禁 PermAvatar 权限位（40302）。
	userGroup.Put("/avatar", func(c fiber.Ctx) error {
		return profile.UpdateAvatarHandler(c, servants)
	})
	// 获取任意用户头像文件 UUID（公开可见，不受拉黑/屏蔽影响）
	userGroup.Get("/users/:uid/avatar", func(c fiber.Ctx) error {
		return profile.GetAvatarHandler(c, servants)
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
	userGroup.Post("/follows", middleware.RequireNotPunished(servants, consts.PermSocial),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{BodyField: "target", CheckMuteUnwatch: true}),
		func(c fiber.Ctx) error {
			return userinteract.FollowHandler(c, servants)
		})
	// 取消关注（幂等）。校验与关注一致。
	userGroup.Delete("/follows", middleware.RequireNotPunished(servants, consts.PermSocial),
		middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{BodyField: "target", CheckMuteUnwatch: true}),
		func(c fiber.Ctx) error {
			return userinteract.UnfollowHandler(c, servants)
		})
	userGroup.Get("/punishments", func(c fiber.Ctx) error {
		return userpunishment.ListMineHandler(c, servants)
	})
	userGroup.Post("/blocks", func(c fiber.Ctx) error {
		return userblock.UpdateHandler(c, servants)
	})
	userGroup.Get("/users/:uid/contents/count", middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{PathParam: "uid"}),
		func(c fiber.Ctx) error {
			return usercontent.CountHandler(c, servants)
		})
	userGroup.Get("/users/:uid/contents", middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{PathParam: "uid"}),
		func(c fiber.Ctx) error {
			return usercontent.ListHandler(c, servants)
		})
	userGroup.Get("/users/:uid/info", func(c fiber.Ctx) error {
		return profile.GetUserInfoHandler(c, servants)
	})
	userGroup.Get("/me", func(c fiber.Ctx) error {
		return profile.MeHandler(c, servants)
	})
	userGroup.Get("/users/:uid/following", middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{PathParam: "uid"}),
		func(c fiber.Ctx) error {
			return userrelation.FollowingHandler(c, servants)
		})
	userGroup.Get("/users/:uid/followers", middleware.RequireNoBlockUser(servants, middleware.BlockGuardConfig{PathParam: "uid"}),
		func(c fiber.Ctx) error {
			return userrelation.FollowersHandler(c, servants)
		})
}
