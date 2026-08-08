package pagination

import (
	"strconv"
	"strings"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Parse 解析 offset/limit 查询参数。limit 缺省 DefaultLimit，最大 MaxLimit；参数非法时 ok=false。
func Parse(offsetRaw, limitRaw string) (offset, limit int, ok bool) {
	offset = 0
	limit = DefaultLimit

	if raw := strings.TrimSpace(offsetRaw); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, false
		}
		offset = value
	}
	if raw := strings.TrimSpace(limitRaw); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > MaxLimit {
			return 0, 0, false
		}
		limit = value
	}
	return offset, limit, true
}
