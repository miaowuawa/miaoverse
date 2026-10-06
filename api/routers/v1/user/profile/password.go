package profile

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"miaoverse/api/routers/v1/sms"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/user/passwordreq"
	"miaoverse/model/server"
	"miaoverse/service/Captcha"
	"miaoverse/service/UserPassword"
	"miaoverse/service/UserSession"
	"miaoverse/service/i18n"
	"miaoverse/util"
)

// SendPasswordSMSHandler 向当前登录账号绑定的手机号发送「修改密码」验证码。
//
// 与公开短信接口的区别：
//   - 手机号与区号取自登录会话，请求体不接受手机号参数，因此无法被用来给任意号码发短信；
//   - 返回的 code_uuid 必须配合 CHANGE_PASSWORD 场景使用，登录验证码不能替代。
func SendPasswordSMSHandler(ctx fiber.Ctx, servants *server.Servants) error {
	// 要求 JSON 请求：跨站表单只能提交 urlencoded / multipart / text-plain，
	// 这一层校验可以挡住「用 CSRF 触发短信发送」的短信轰炸路径。
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	phone, region, ok := UserSession.CurrentPhoneRegion(ctx)
	if !ok {
		return resp.PhoneNotBound(ctx)
	}

	regionText := strconv.FormatUint(uint64(region), 10)
	codeUUID, err := Captcha.Send(
		servants.CodeManager,
		servants.SmsServant,
		consts.ActionChangePassword,
		regionText,
		phone,
		i18n.Message(ctx, i18n.SMSActionChangePassword),
	)
	return sms.SmsResult(ctx, codeUUID, err)
}

// PasswordStatusHandler 返回当前登录账号是否已设置密码，供前端展示「设置密码 / 修改密码」。
// 只允许查询本人（uid 取自登录会话），不接收任何用户标识参数。
func PasswordStatusHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	hasPassword, err := servants.UserServant.HasCredential(uid, consts.Password)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.PasswordStatus(ctx, hasPassword)
}

// UpdatePasswordHandler 通过手机验证码设置/修改当前登录账号的密码。
//
// 安全约束：
//   - 手机号与区号一律取自当前登录会话，请求体不接受手机号参数，因此不存在
//     通过改包修改他人账号密码的越权路径（也不支持把密码改绑到其他手机号）；
//   - 验证码必须是 CHANGE_PASSWORD 场景下发的（登录验证码无法复用），一次性且在 Redis 中限时有效；
//   - 密码强度校验不通过时直接返回，不消耗一次性验证码；
//   - 明文密码只在内存中用于 bcrypt 哈希，响应与日志均不回显；
//   - 修改成功后重新生成 session ID，防御会话固定。
func UpdatePasswordHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	// 手机号来自会话：仅短信登录会话写入了手机号；没有手机号说明账号未绑定手机号，无法用短信改密。
	phone, region, ok := UserSession.CurrentPhoneRegion(ctx)
	if !ok {
		return resp.PhoneNotBound(ctx)
	}

	req := &passwordreq.ChangePassword{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	if err := servants.Validator.Struct(req); err != nil {
		return resp.BadRequest(ctx)
	}

	valid, _ := util.Security.ValidateAvalue(req.Timestamp)
	if !valid {
		return ctx.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
			Code: fiber.StatusBadRequest,
			Msg:  i18n.Message(ctx, i18n.ErrRequestTimeout),
		})
	}

	// 先校验密码强度再校验验证码：弱密码直接驳回，避免白白消耗一次性验证码。
	if !UserPassword.ValidateStrength(req.Password) {
		return ctx.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
			Code: fiber.StatusBadRequest,
			Msg:  i18n.Message(ctx, i18n.ErrPasswordTooWeak),
		})
	}

	passed, err := servants.CodeManager.VerifySceneCode(
		consts.ActionChangePassword,
		strconv.FormatUint(uint64(region), 10),
		phone,
		req.UUID,
		strconv.Itoa(req.Code),
	)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if !passed {
		return ctx.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  i18n.Message(ctx, i18n.ErrSMSCodeInvalid),
		})
	}

	if err := UserPassword.SetPassword(servants.UserServant, uid, req.Password); err != nil {
		return resp.ServerError(ctx)
	}

	// 密码修改成功后向账号本人发送「账号安全」通知（旁路业务，失败不影响修改结果）
	servants.NotifyServant.NotifyAccountSecurity(uid, i18n.Message(ctx, i18n.NotifyPasswordChanged))

	if err := UserSession.RefreshSessionID(ctx); err != nil {
		return resp.ServerError(ctx)
	}

	return resp.PasswordUpdated(ctx)
}
