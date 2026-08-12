package sms

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/resp/smsresp"
	"miaoverse/model/dto/smsreq"
	"miaoverse/model/server"
	"miaoverse/service/i18n"
	"miaoverse/util"
)

func SendSmsHandler(c fiber.Ctx, servants *server.Servants) error {
	if !c.IsJSON() {
		return resp.BadRequest(c)
	}

	req := &smsreq.GetSmsReq{}
	if err := c.Bind().Body(req); err != nil {
		return resp.BadRequest(c)
	}
	if err := servants.Validator.Struct(req); err != nil {
		return resp.BadRequest(c)
	}

	valid, _ := util.Security.ValidateAvalue(req.Timestamp)
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
		// 不向客户端回显短信服务内部错误（可能包含账号/余额等敏感状态），统一返回通用文案
		return c.Status(fiber.StatusInternalServerError).JSON(resp.CodeWithMsg{
			Code: fiber.StatusInternalServerError,
			Msg:  i18n.Message(c, i18n.ErrSMSProvider),
		})
	}

	return c.Status(fiber.StatusOK).JSON(smsresp.SmsResp{
		CodeUUID: codeUUID,
		Msg:      i18n.Message(c, i18n.OKSMSSent),
	})
}
