package sms

import (
	"miaoverse/service/security/sms/codemanager"

	"github.com/gofiber/fiber/v3"
)

func ValidateSms(c fiber.Ctx, codeManager *codemanager.CodeManager) {

	codeManager.VerifyCodeByPhoneMD5()
}
