package validate

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

// 1. 预编译 UUID v4 正则表达式（只编译一次，提升性能）
var uuidV4Regex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// 2. 自定义 UUID v4 验证函数
func isUUIDV4(fl validator.FieldLevel) bool {
	uuid := fl.Field().String()
	// 匹配 UUID v4 格式（核心：第4段以 4 开头，第5段以 8/9/a/b 开头）
	return uuidV4Regex.MatchString(uuid)
}
