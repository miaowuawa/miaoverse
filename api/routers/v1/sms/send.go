package sms

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/resp/smsresp"
	"miaoverse/model/dto/smsreq"
	"miaoverse/model/server"
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
			Msg:  "请求超时，请重新尝试",
		})
	}

	err, code, codeUUID := servants.CodeManager.PrepareCodeForPhone(req.Region, req.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(resp.CodeWithMsg{
			Code: fiber.StatusInternalServerError,
			Msg:  "服务器内部错误，请稍后重试",
		})
	}

	err = servants.SmsServant.SendPhoneCaptcha(req.Phone, code, time.Minute*5, "登录或注册")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(resp.CodeWithMsg{
			Code: fiber.StatusInternalServerError,
			Msg:  "返回异常：" + err.Error() + "请联系管理人员！",
		})
	}

	return c.Status(fiber.StatusOK).JSON(smsresp.SmsResp{
		CodeUUID: codeUUID,
		Msg:      "发送成功",
	})
}

func badRequest(c fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
		Code: fiber.StatusBadRequest,
		Msg:  "请求错误，请检查参数",
	})
}
