package sms

import (
	"github.com/gofiber/fiber/v3"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/resp/smsresp"
	"miaoverse/model/dto/smsreq"
	"miaoverse/service/security/sms"
	"miaoverse/service/security/sms/codemanager"
	"miaoverse/service/security/sms/smsbao"
	"time"
)

func SendSmsHandler(c fiber.Ctx, codeManager *codemanager.CodeManager, smsBaoServant *smsbao.SmsBaoServant) error {
	if !c.IsJSON() {
		_ = c.Status(fiber.StatusBadRequest).JSON(
			resp.CodeWithMsg{
				Code: fiber.StatusBadRequest,
				Msg:  "请求错误，请升级最新版本喵星",
			},
		)
	}
	req := &smsreq.GetSmsReq{}
	err := c.Bind().Body(req)
	if err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(
			resp.CodeWithMsg{
				Code: fiber.StatusBadRequest,
				Msg:  "请求错误，请升级最新版本喵星",
			},
		)
	} //处理反序列化错误
	// validate A value
	valid, _ := sms.ValidateAvalue(req.Timestamp)
	if !valid {
		_ = c.Status(fiber.StatusBadRequest).JSON(
			resp.CodeWithMsg{
				Code: fiber.StatusBadRequest,
				Msg:  "请求超时，请重新尝试",
			},
		)
	} //处理A校验错误
	err, codeID, codeUUID := codeManager.PrepareCodeForPhone(req.Region, req.Phone)
	if err != nil {
		_ = c.Status(fiber.StatusInternalServerError).JSON(
			resp.CodeWithMsg{
				Code: fiber.StatusInternalServerError,
				Msg:  "服务器内部错误，请稍后重试",
			},
		)
	} //处理preparecode错误

	err = smsBaoServant.SendPhoneCaptcha(req.Phone, codeID, time.Minute*5)
	if err != nil {
		_ = c.Status(fiber.StatusInternalServerError).JSON(
			resp.CodeWithMsg{
				Code: fiber.StatusInternalServerError,
				Msg:  "返回异常：" + err.Error() + "请联系管理人员！",
			},
		)
	}

	response := smsresp.SmsResp{
		CodeId:   codeID,
		CodeUuid: codeUUID,
		Msg:      "发送成功",
	}

	_ = c.Status(fiber.StatusInternalServerError).JSON(
		response,
	)
	return nil
}
