package buildstring

import "fmt"

// GenerateUsername 生成默认格式的用户名：miaoverse_8位字母数字混合
// 参数说明：
//
//	ctx: 上下文（用于Redis操作）
//	redisClient: Redis客户端（可选，传nil则不做重复检测）
//
// 返回值：
//
//	string: 生成的用户名
//	error: 生成失败/重复检测失败的错误
func GenerateUsername() (string, error) {
	// 1. 生成8位随机字母数字混合字符串
	randomStr, _ := generateRandomString(8)

	// 2. 拼接默认用户名格式
	username := fmt.Sprintf("miaoverse_%s", randomStr)
	return username, nil
}
