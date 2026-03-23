package maths

import "math/rand"

// RandomIntLimited 生成范围内随机数
func RandomIntLimited(min int, max int) int {
	return rand.Intn(max-min) + min
}
