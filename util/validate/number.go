package validate

import (
	"github.com/go-playground/validator/v10"
	"unicode"
)

// isDigit  验证字符串是否为数字
func isDigit(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return false
	}
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}
