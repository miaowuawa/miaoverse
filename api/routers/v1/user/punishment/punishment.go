package punishment

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/middleware"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
)

// ListMineHandler 查询当前登录用户本人的全部惩罚记录
func ListMineHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	list, err := servants.UserServant.QueryPunishmentsByUser(uid)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if list == nil {
		list = []modeluser.Punishment{}
	}

	return resp.PunishmentsList(ctx, list)
}
