package codemanager

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

// VerifyCodeByRegionPhoneMD5 验证验证码是否正确
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
func (c *CodeManager) VerifyCodeByRegionPhoneMD5(regionPhoneMD5 string, codeUUID string, inputCode string) (bool, error) {
	// 1. 拼接codeID（和生成时的规则保持一致）
	codeID := codeUUID + "-" + regionPhoneMD5

	existsCmd := c.Redis.Exists(c.Context, codeID)
	// 第一步：获取执行结果和错误
	count, err := existsCmd.Result()
	// 第二步：先处理命令执行错误（比如 Redis 连不上）
	if err != nil {
		return false, fmt.Errorf("检查验证码存在性失败：%w", err)
	}
	// 第三步：判断 Key 是否不存在（count == 0 表示不存在）
	if count == 0 {
		return false, errors.New("验证码不存在")
	}

	// 2. 从Redis中获取存储的验证码
	result := c.Redis.Get(c.Context, codeID)
	if result.Err() != nil {
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

	if delErr := c.Redis.Del(c.Context, codeID).Err(); delErr != nil {
		// 删除失败就算验证不过，避免重用验证码
		log.Printf("删除验证码失败，codeID=%s, err=%v", codeID, delErr)
		return false, errors.New("验证码验证成功，但删除失败: " + delErr.Error())
	}

	return true, nil
}
