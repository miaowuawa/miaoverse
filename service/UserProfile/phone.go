package UserProfile

import (
	"strconv"
	"strings"
)

// FormatMaskedPhone 把绑定手机号格式化为「+86 138****8000」形式，用于个人资料页只读展示。
//
// 打码而不是返回完整号码：本人资料页只需要让用户确认绑定的是哪个号，
// 完整号码一旦进入响应体，就可能被截图、日志或前端状态持久化带出。
// 手机号当前不支持更改，因此这里只做展示，不提供任何写入路径。
func FormatMaskedPhone(phone string, region uint16) string {
	masked := MaskPhone(phone)
	if masked == "" {
		return ""
	}
	if region == 0 {
		return masked
	}
	return "+" + strconv.FormatUint(uint64(region), 10) + " " + masked
}

// MaskPhone 保留手机号前 3 位与后 4 位，中间统一用 * 代替；
// 位数不足 8 位时整体打码，避免短号码被打码后仍可还原。
func MaskPhone(phone string) string {
	digits := []rune(strings.TrimSpace(phone))
	if len(digits) == 0 {
		return ""
	}
	if len(digits) < 8 {
		return strings.Repeat("*", len(digits))
	}
	return string(digits[:3]) + strings.Repeat("*", len(digits)-7) + string(digits[len(digits)-4:])
}
