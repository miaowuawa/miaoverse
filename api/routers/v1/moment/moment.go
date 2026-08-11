package moment

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/middleware"
	"miaoverse/model/dto/moment/publishreq"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/UserMoment"
)

func PublishHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &publishreq.PublishMoment{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}

	record, ok := UserMoment.NormalizePublish(req)
	if !ok {
		return resp.BadRequest(ctx)
	}
	record.UserID = uid

	created, err := servants.ContentServant.CreateMoment(*record)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.MomentPublished(ctx, UserMoment.ToMomentInfo(created))
}

// DetailHandler 获取动态详情。拉黑/被拉黑校验由 RequireNoBlock 中间件完成（40301），
// 动态存在性与状态校验由 ResolveMomentPathAuthor 完成（404），
// 此处仅做可见性（permission）校验与详情组装。
func DetailHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
	}

	isFriend, isFan, err := UserMoment.RelationFlags(servants.InteractsServant, uid, moment.UserID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if !UserMoment.VisibleTo(moment, uid, isFriend, isFan) {
		return resp.FileNotFound(ctx)
	}

	// is_following 表示当前用户是否关注了作者（单向关注），与可见性判定的互相关注口径不同
	isFollowing, err := servants.InteractsServant.IsFollowing(uid, moment.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return resp.ServerError(ctx)
	}

	author, err := servants.UserServant.QueryByID(moment.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.FileNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}

	meta, err := servants.ContentServant.QueryMomentInteractCount(moment.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return resp.ServerError(ctx)
	}

	isLiked, err := servants.InteractsServant.HasLikedMoment(uid, moment.ID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.MomentDetailOK(ctx, UserMoment.ToMomentDetail(moment, author, meta, isLiked, isFollowing))
}
