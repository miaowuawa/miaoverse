package validate

import "github.com/go-playground/validator/v10"

// isDigit 验证字符串是否仅包含 ASCII 数字（0-9）。
// 不使用 unicode.IsDigit：它会放行全角数字、阿拉伯-印度数字等非 ASCII 数字字符。
func isDigit(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isPhone 验证手机号：仅允许 5-15 位纯 ASCII 数字（不含区号，符合 E.164 国家码规范）。
func isPhone(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) < 5 || len(s) > 15 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
