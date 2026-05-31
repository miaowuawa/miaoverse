package sms

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/resp/smsresp"
	"miaoverse/model/dto/smsreq"
	"miaoverse/model/server"
	"miaoverse/service/i18n"
	"miaoverse/service/security"
)

func SendSmsHandler(c fiber.Ctx, servants *server.Servants) error {
	if !c.IsJSON() {
		return badRequest(c)
	}

	req := &smsreq.GetSmsReq{}
	if err := c.Bind().Body(req); err != nil {
		return badRequest(c)
	}
	if err := servants.Validator.Struct(req); err != nil {
		return badRequest(c)
	}

	valid, _ := security.ValidateAvalue(req.Timestamp)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
			Code: fiber.StatusBadRequest,
			Msg:  i18n.Message(c, i18n.ErrRequestTimeout),
		})
	}

	err, code, codeUUID := servants.CodeManager.PrepareCodeForPhone(req.Region, req.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(resp.CodeWithMsg{
			Code: fiber.StatusInternalServerError,
			Msg:  i18n.Message(c, i18n.ErrServerInternal),
		})
	}

	err = servants.SmsServant.SendPhoneCaptcha(req.Phone, code, time.Minute*5, i18n.Message(c, i18n.SMSActionLoginRegister))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(resp.CodeWithMsg{
			Code: fiber.StatusInternalServerError,
			Msg:  i18n.T(c, i18n.ErrSMSProvider, i18n.Data{"error": err.Error()}),
		})
	}

	return c.Status(fiber.StatusOK).JSON(smsresp.SmsResp{
		CodeUUID: codeUUID,
		Msg:      i18n.Message(c, i18n.OKSMSSent),
	})
}

func badRequest(c fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
		Code: fiber.StatusBadRequest,
		Msg:  i18n.Message(c, i18n.ErrBadRequest),
	})
}
