package codemanager

import (
	"github.com/gofiber/utils/v2"
	"miaoverse/util/encrypt/md5hash"
	"miaoverse/util/maths"
)

func generateCode() int {
	return maths.RandomIntLimited(100000, 999999)
}

// GetCodeForPhone 为一个手机号准备一个验证码，返回成功/失败河验证码uuid
func (c *CodeManager) GetCodeForPhone(region string, phone string) (error, string, string) {
	codeUUID := utils.UUID()
	codeID := codeUUID + md5hash.HashStr(region+phone)
	//TODO:启动事务
	result := c.Redis.Set(c.Context, codeID, generateCode(), c.CodeExpireTime)
	if result.Err() != nil {
		return result.Err(), "", ""
	}
	return nil, codeID, codeUUID
}
