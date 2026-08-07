package resp

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/model/dao/user"
	"miaoverse/service/i18n"
)

type CodeWithMsgPunishments struct {
	Code        int               `json:"code"`
	Msg         string            `json:"msg"`
	Punishments []user.Punishment `json:"punishments"`
}

func PunishmentsList(ctx fiber.Ctx, list []user.Punishment) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgPunishments{
		Code:        fiber.StatusOK,
		Msg:         i18n.Message(ctx, i18n.OKPunishmentsFetched),
		Punishments: list,
	})
}
