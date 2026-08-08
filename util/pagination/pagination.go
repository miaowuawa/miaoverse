package pagination

import (
	"strconv"
	"strings"

	"miaoverse/consts"
)

// Parse 解析 offset/limit 查询参数。limit 缺省 consts.DefaultListLimit，最大 consts.MaxListLimit；参数非法时 ok=false。
func Parse(offsetRaw, limitRaw string) (offset, limit int, ok bool) {
	offset = 0
	limit = consts.DefaultListLimit

	if raw := strings.TrimSpace(offsetRaw); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, false
		}
		offset = value
	}
	if raw := strings.TrimSpace(limitRaw); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > consts.MaxListLimit {
			return 0, 0, false
		}
		limit = value
	}
	return offset, limit, true
}
