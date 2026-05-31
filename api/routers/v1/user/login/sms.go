package login

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/user/loginreq"
	"miaoverse/model/server"
	"miaoverse/service/UserCheck"
	"miaoverse/service/UserSession"
	"miaoverse/service/security"
	"miaoverse/util/encrypt/md5hash"
)

func BySMSHandler(ctx fiber.Ctx, servants *server.Servants) error {
	req, ok := bindAndValidateSMS(ctx, servants)
	if !ok {
		return badRequest(ctx)
	}

	if ok, err := verifySMSCode(req, servants); err != nil {
		return serverError(ctx)
	} else if !ok {
		return ctx.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  "验证码错误或不存在，请重试",
		})
	}

	users, exists, err := UserCheck.CheckAndCreateIfNotExists(req.Phone, req.Region, servants)
	if err != nil {
		return serverError(ctx)
	}
	if len(users) == 0 {
		return serverError(ctx)
	}

	if !exists {
		return loginSingleAccount(ctx, req.Phone, req.Region, users[0].ID, fiber.StatusCreated, "注册并登录成功")
	}

	if len(users) == 1 {
		return loginSingleAccount(ctx, req.Phone, req.Region, users[0].ID, fiber.StatusOK, "登录成功")
	}

	if err := UserSession.LoginBySMSMultipleChoices(ctx, req.Phone, req.Region); err != nil {
		return serverError(ctx)
	}
	return ctx.Status(fiber.StatusMultipleChoices).JSON(resp.CodeWithMsgUserChoice{
		Code:    fiber.StatusMultipleChoices,
		Msg:     "请选择要登录的账号",
		Choices: users,
	})
}

func RegisterBySMSHandler(ctx fiber.Ctx, servants *server.Servants) error {
	req, ok := bindAndValidateSMS(ctx, servants)
	if !ok {
		return badRequest(ctx)
	}

	if ok, err := verifySMSCode(req, servants); err != nil {
		return serverError(ctx)
	} else if !ok {
		return ctx.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  "验证码错误或不存在，请重试",
		})
	}

	users, err := servants.UserServant.QueryByPhone(req.Phone, req.Region)
	if err != nil {
		return serverError(ctx)
	}
	if len(users) == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(resp.CodeWithMsg{
			Code: fiber.StatusNotFound,
			Msg:  "该手机号还没有账号，请使用短信登录自动注册首个账号",
		})
	}

	newUser, err := UserCheck.CreateAccountForPhone(req.Phone, req.Region, servants)
	if err != nil {
		return serverError(ctx)
	}

	return loginSingleAccount(ctx, req.Phone, req.Region, newUser.ID, fiber.StatusCreated, "新账号注册并登录成功")
}

func ChooseUserHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if !ctx.IsJSON() {
		return badRequest(ctx)
	}

	req := &loginreq.ChooseAccount{}
	if err := ctx.Bind().Body(req); err != nil {
		return badRequest(ctx)
	}
	if err := servants.Validator.Struct(req); err != nil {
		return badRequest(ctx)
	}

	phone, region, ok := UserSession.PendingSMSLogin(ctx)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
			Code: fiber.StatusBadRequest,
			Msg:  "没有待选择的登录账号，请重新验证码登录",
		})
	}

	belongs, err := UserCheck.UserBelongsToPhone(req.UID, phone, region, servants)
	if err != nil {
		return serverError(ctx)
	}
	if !belongs {
		return ctx.Status(fiber.StatusForbidden).JSON(resp.CodeWithMsg{
			Code: fiber.StatusForbidden,
			Msg:  "账号不属于本次验证的手机号",
		})
	}

	return loginSingleAccount(ctx, phone, region, req.UID, fiber.StatusOK, "登录成功")
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
	valid, _ := security.ValidateAvalue(req.TimeStamp)
	if !valid {
		return false, nil
	}
	hash := md5hash.HashStr(strconv.Itoa(req.Region) + req.Phone)
	return servants.CodeManager.VerifyCodeByRegionPhoneMD5(hash, req.UUID, strconv.Itoa(req.Code))
}

func loginSingleAccount(ctx fiber.Ctx, phone string, region int, uid uint64, status int, msg string) error {
	if err := UserSession.LoginBySMSSingleAccount(ctx, phone, region, uid); err != nil {
		return serverError(ctx)
	}
	return ctx.Status(status).JSON(resp.CodeWithMsgUserID{
		Code: status,
		Msg:  msg,
		UID:  uid,
	})
}

func badRequest(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
		Code: fiber.StatusBadRequest,
		Msg:  "请求错误，请检查参数",
	})
}

func serverError(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusInternalServerError).JSON(resp.CodeWithMsg{
		Code: fiber.StatusInternalServerError,
		Msg:  "服务器异常，请联系管理员",
	})
}
