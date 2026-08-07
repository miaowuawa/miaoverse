package moment

import (
	"github.com/gofiber/fiber/v3"
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
