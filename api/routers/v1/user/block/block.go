package block

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/user/blockreq"
	"miaoverse/model/server"
	"miaoverse/service/UserBlock"
)

func UpdateHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &blockreq.UpdateBlock{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}

	action := strings.TrimSpace(req.Action)
	blockType := UserBlock.BlockType(req.Type)
	if req.Target == 0 || req.Target == uid {
		return resp.BadRequest(ctx)
	}
	if blockType != UserBlock.BlockTypeBlock && blockType != UserBlock.BlockTypeMute && blockType != UserBlock.BlockTypeUnwatch {
		return resp.BadRequest(ctx)
	}
	if action != consts.ActionAdd && action != consts.ActionRemove {
		return resp.BadRequest(ctx)
	}

	if action == consts.ActionAdd {
		if err := servants.BlockServant.Add(ctx.Context(), uid, blockType, req.Target); err != nil {
			return resp.ServerError(ctx)
		}
	} else {
		if err := servants.BlockServant.Remove(ctx.Context(), uid, blockType, req.Target); err != nil {
			return resp.ServerError(ctx)
		}
	}

	return resp.BlockUpdated(ctx, req.Target, req.Type, action)
}
