package login

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/user/loginreq"
	"miaoverse/model/server"
	"miaoverse/service/UserCheck"
	"miaoverse/service/UserSession"
	"miaoverse/service/i18n"
	"miaoverse/util"
)

func BySMSHandler(ctx fiber.Ctx, servants *server.Servants) error {
	req, ok := bindAndValidateSMS(ctx, servants)
	if !ok {
		return resp.BadRequest(ctx)
	}

	if ok, err := verifySMSCode(req, servants); err != nil {
		return resp.ServerError(ctx)
	} else if !ok {
		return ctx.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  i18n.Message(ctx, i18n.ErrSMSCodeInvalid),
		})
	}

	users, exists, err := UserCheck.CheckAndCreateIfNotExists(req.Phone, req.Region, servants)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if len(users) == 0 {
		return resp.ServerError(ctx)
	}

	if !exists {
		return loginSingleAccount(ctx, req.Phone, req.Region, users[0].ID, fiber.StatusCreated, i18n.OKRegisterAndLogin)
	}

	if len(users) == 1 {
		return loginSingleAccount(ctx, req.Phone, req.Region, users[0].ID, fiber.StatusOK, i18n.OKLogin)
	}

	if err := UserSession.LoginBySMSMultipleChoices(ctx, req.Phone, req.Region); err != nil {
		return resp.ServerError(ctx)
	}
	return ctx.Status(fiber.StatusMultipleChoices).JSON(resp.CodeWithMsgUserChoice{
		Code:    fiber.StatusMultipleChoices,
		Msg:     i18n.Message(ctx, i18n.OKChooseLoginAccount),
		Choices: users,
	})
}

func RegisterBySMSHandler(ctx fiber.Ctx, servants *server.Servants) error {
	req, ok := bindAndValidateSMS(ctx, servants)
	if !ok {
		return resp.BadRequest(ctx)
	}

	if ok, err := verifySMSCode(req, servants); err != nil {
		return resp.ServerError(ctx)
	} else if !ok {
		return ctx.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  i18n.Message(ctx, i18n.ErrSMSCodeInvalid),
		})
	}

	users, err := servants.UserServant.QueryByPhone(req.Phone, req.Region)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if len(users) == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(resp.CodeWithMsg{
			Code: fiber.StatusNotFound,
			Msg:  i18n.Message(ctx, i18n.ErrPhoneHasNoAccount),
		})
	}

	newUser, err := UserCheck.CreateAccountForPhone(req.Phone, req.Region, servants)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return loginSingleAccount(ctx, req.Phone, req.Region, newUser.ID, fiber.StatusCreated, i18n.OKNewAccountAndLogin)
}

func ChooseUserHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &loginreq.ChooseAccount{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	if err := servants.Validator.Struct(req); err != nil {
		return resp.BadRequest(ctx)
	}

	phone, region, ok := UserSession.PendingSMSLogin(ctx)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
			Code: fiber.StatusBadRequest,
			Msg:  i18n.Message(ctx, i18n.ErrNoPendingLoginAccount),
		})
	}

	belongs, err := UserCheck.UserBelongsToPhone(req.UID, phone, region, servants)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if !belongs {
		return ctx.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  i18n.Message(ctx, i18n.ErrAccountNotBelongPhone),
		})
	}

	return loginSingleAccount(ctx, phone, region, req.UID, fiber.StatusOK, i18n.OKLogin)
}

// AccountListHandler 返回当前会话登录手机号绑定的全部可登录账号。
// 仅要求已登录（不校验当前账号状态），被封禁的账号也可以看到账号列表并切换到其他正常账号。
func AccountListHandler(ctx fiber.Ctx, servants *server.Servants) error {
	phone, region, ok := UserSession.CurrentPhoneRegion(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	users, err := servants.UserServant.QueryByPhoneLoginable(phone, region)
	if err != nil {
		return resp.ServerError(ctx)
	}

	current, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	return ctx.Status(fiber.StatusOK).JSON(resp.CodeWithMsgAccountList{
		Code:    fiber.StatusOK,
		Msg:     i18n.Message(ctx, i18n.OKAccountList),
		Current: current,
		Users:   users,
	})
}

// SwitchAccountHandler 将当前会话切换到同一手机号绑定的另一个账号。
// 仅要求已登录（不校验当前账号状态）；目标账号必须属于当前会话手机号且未被封禁。
func SwitchAccountHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &loginreq.SwitchAccount{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	if err := servants.Validator.Struct(req); err != nil {
		return resp.BadRequest(ctx)
	}

	phone, region, ok := UserSession.CurrentPhoneRegion(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	belongs, err := UserCheck.UserBelongsToPhone(req.UID, phone, region, servants)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if !belongs {
		return ctx.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  i18n.Message(ctx, i18n.ErrAccountNotBelongPhone),
		})
	}

	user, err := servants.UserServant.QueryByID(req.UID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if user.Status == consts.UserStatusBanned {
		return resp.AccountBanned(ctx)
	}

	return loginSingleAccount(ctx, phone, region, req.UID, fiber.StatusOK, i18n.OKLogin)
}

func bindAndValidateSMS(ctx fiber.Ctx, servants *server.Servants) (*loginreq.SMS, bool) {
	if !ctx.IsJSON() {
		return nil, false
	}
	req := &loginreq.SMS{}
	if err := ctx.Bind().Body(req); err != nil {
		return nil, false
	}
	if err := servants.Validator.Struct(req); err != nil {
		return nil, false
	}
	return req, true
}

func verifySMSCode(req *loginreq.SMS, servants *server.Servants) (bool, error) {
	valid, _ := util.Security.ValidateAvalue(req.TimeStamp)
	if !valid {
		return false, nil
	}
	hash := util.MD5Hash.HashStr(strconv.FormatUint(uint64(req.Region), 10) + req.Phone)
	return servants.CodeManager.VerifyCodeByRegionPhoneMD5(hash, req.UUID, strconv.Itoa(req.Code))
}

func loginSingleAccount(ctx fiber.Ctx, phone string, region uint16, uid uint32, status int, msgKey i18n.MessageKey) error {
	if err := UserSession.LoginBySMSSingleAccount(ctx, phone, region, uid); err != nil {
		return resp.ServerError(ctx)
	}
	return ctx.Status(status).JSON(resp.CodeWithMsgUserID{
		Code: status,
		Msg:  i18n.Message(ctx, msgKey),
		UID:  uid,
	})
}

// LogoutHandler 退出登录：销毁服务端 session 并清除登录态 cookie。
func LogoutHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if err := UserSession.Logout(ctx); err != nil {
		return resp.ServerError(ctx)
	}
	return ctx.Status(fiber.StatusOK).JSON(resp.CodeWithMsg{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKLogout),
	})
}
