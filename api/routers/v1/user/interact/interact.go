package interact

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/i18n"
)

// FollowHandler 关注用户。拉黑/屏蔽/不想看校验由 RequireNoBlockUser 中间件完成。
func FollowHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	targetID, ok := middleware.BlockTarget(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	if err := servants.InteractsServant.FollowUser(uid, targetID); err != nil {
		return resp.ServerError(ctx)
	}

	// 被关注通知（旁路业务，失败不影响关注结果）
	if actor, ok := middleware.CurrentUser(ctx); ok {
		servants.NotifyServant.NotifyFollow(actor, targetID, i18n.LanguageFromCtx(ctx))
	}

	return resp.InteractOK(ctx, uint64(targetID), consts.ActionFollow)
}

// UnfollowHandler 取消关注（幂等：未关注时直接返回成功）。
// 拉黑/屏蔽/不想看校验与关注一致（RequireNoBlockUser 中间件）。
func UnfollowHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	targetID, ok := middleware.BlockTarget(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	if err := servants.InteractsServant.UnfollowUser(uid, targetID); err != nil {
		return resp.ServerError(ctx)
	}

	return resp.InteractOK(ctx, uint64(targetID), consts.ActionUnfollow)
}

// LikeHandler 给动态点赞。内容屏蔽校验由 RequireNoContentBlock、拉黑校验由 RequireNoBlockUser 中间件完成。
func LikeHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	if err := servants.InteractsServant.LikeMomentAndMeta(uid, moment.ID, moment.UserID); err != nil {
		return resp.ServerError(ctx)
	}

	// 被点赞通知（旁路业务，失败不影响点赞结果；自己给自己点赞由服务层跳过）
	if actor, ok := middleware.CurrentUser(ctx); ok {
		servants.NotifyServant.NotifyLikeMoment(actor, moment.ID, moment.UserID, moment.Content, i18n.LanguageFromCtx(ctx))
	}

	return resp.InteractOK(ctx, moment.ID, consts.ActionLike)
}

// UnlikeHandler 取消点赞（幂等：未点赞时直接返回成功）。
// 内容屏蔽校验由 RequireNoContentBlock、拉黑校验由 RequireNoBlockUser 中间件完成。
func UnlikeHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	if err := servants.InteractsServant.UnlikeMomentAndMeta(uid, moment.ID); err != nil {
		return resp.ServerError(ctx)
	}

	return resp.InteractOK(ctx, moment.ID, consts.ActionUnlike)
}
