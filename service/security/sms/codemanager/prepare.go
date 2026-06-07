package codemanager

import (
	"strconv"

	"github.com/gofiber/utils/v2"
	"miaoverse/util"
)

func generateCode() int {
	code, _ := util.Maths.RandomIntLimited(1000, 10000)
	return code
}

func (c *CodeManager) PrepareCodeForPhone(region string, phone string) (error, string, string) {
	codeUUID := utils.UUIDv4()
	codeID := codeUUID + "-" + util.MD5Hash.HashStr(region+phone)
	code := generateCode()

	result := c.Redis.Set(c.Context, codeID, code, c.CodeExpireTime)
	if result.Err() != nil {
		return result.Err(), "", ""
	}
	return nil, strconv.Itoa(code), codeUUID
}
