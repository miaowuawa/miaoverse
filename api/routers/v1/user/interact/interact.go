package interact

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
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

	if err := servants.InteractsServant.LikeMomentAndMeta(uid, moment.ID); err != nil {
		return resp.ServerError(ctx)
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
