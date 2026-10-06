package sms

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/resp/smsresp"
	"miaoverse/model/dto/smsreq"
	"miaoverse/model/server"
	"miaoverse/service/Captcha"
	"miaoverse/service/i18n"
	"miaoverse/service/security/sms/codemanager"
	"miaoverse/util"
)

// SendSmsHandler 公开的短信验证码申请接口（无需登录）。
// 只允许公开场景（见 consts.PublicSMSActions）：修改密码验证码必须走登录态的
// POST /api/v1/user/password/sms，手机号由服务端从会话取，避免被用来轰炸任意号码。
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

	// 业务场景白名单：不传按登录/注册处理；非公开场景直接拒绝。
	action := strings.ToUpper(strings.TrimSpace(req.Action))
	if action == "" {
		action = consts.ActionLogin
	}
	if !consts.IsPublicSMSAction(action) {
		return resp.BadRequest(c)
	}

	valid, _ := util.Security.ValidateAvalue(req.Timestamp)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
			Code: fiber.StatusBadRequest,
			Msg:  i18n.Message(c, i18n.ErrRequestTimeout),
		})
	}

	codeUUID, err := Captcha.Send(
		servants.CodeManager,
		servants.SmsServant,
		action,
		req.Region,
		req.Phone,
		i18n.Message(c, i18n.SMSActionLoginRegister),
	)
	return SmsResult(c, codeUUID, err)
}

// SmsResult 把验证码发送结果统一映射为 HTTP 响应，供公开短信接口与
// 登录态修改密码短信接口复用，避免两处各写一套状态码映射。
// 短信网关错误不向客户端回显（可能包含账号、余额等敏感状态）。
func SmsResult(c fiber.Ctx, codeUUID string, err error) error {
	if errors.Is(err, codemanager.ErrTooFrequent) {
		// 冷却期内重复申请：返回 429，客户端应等待倒计时结束
		return c.Status(fiber.StatusTooManyRequests).JSON(resp.CodeWithMsg{
			Code: fiber.StatusTooManyRequests,
			Msg:  i18n.Message(c, i18n.ErrSMSTooFrequent),
		})
	}
	if errors.Is(err, Captcha.ErrProvider) {
		return c.Status(fiber.StatusInternalServerError).JSON(resp.CodeWithMsg{
			Code: fiber.StatusInternalServerError,
			Msg:  i18n.Message(c, i18n.ErrSMSProvider),
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(resp.CodeWithMsg{
			Code: fiber.StatusInternalServerError,
			Msg:  i18n.Message(c, i18n.ErrServerInternal),
		})
	}

	return c.Status(fiber.StatusOK).JSON(smsresp.SmsResp{
		CodeUUID: codeUUID,
		Msg:      i18n.Message(c, i18n.OKSMSSent),
	})
}
