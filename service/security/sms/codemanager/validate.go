package codemanager

import (
	"errors"
	"strconv"
)

// VerifyCodeByPhoneMD5 验证验证码是否正确
// 参数说明：
//
//	phoneMD5: 手机号的MD5字符串
//	codeUUID: 生成验证码时返回的UUID
//	inputCode: 用户输入的验证码（字符串格式，兼容前端传参）
//
// 返回值：
//
//	bool: 验证是否通过
//	error: 错误信息（Redis操作失败/其他系统错误）
func (c *CodeManager) VerifyCodeByPhoneMD5(phoneMD5 string, codeUUID string, inputCode string) (bool, error) {
	// 1. 拼接codeID（和生成时的规则保持一致）
	codeID := codeUUID + phoneMD5

	// 2. 从Redis中获取存储的验证码
	result := c.Redis.Get(c.Context, codeID)
	if result.Err() != nil {
		// 如果是key不存在，说明验证码已过期或不存在
		if result.Err().Error() == "redis: nil" {
			return false, nil
		}
		// 其他Redis错误
		return false, result.Err()
	}

	// 3. 解析Redis中的验证码（存储的是int类型）
	storeCodeStr := result.Val()
	storeCode, err := strconv.Atoi(storeCodeStr)
	if err != nil {
		return false, errors.New("验证码格式错误: " + err.Error())
	}

	// 4. 解析用户输入的验证码
	inputCodeInt, err := strconv.Atoi(inputCode)
	if err != nil {
		return false, errors.New("输入的验证码格式错误: " + err.Error())
	}

	// 5. 对比验证码
	if storeCode != inputCodeInt {
		return false, nil
	}

	// 6. 验证成功后建议删除验证码（防止重复使用）
	_ = c.Redis.Del(c.Context, codeID).Err()

	return true, nil
}
