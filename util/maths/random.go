package maths

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// RandomIntLimited 生成范围内随机数（使用 crypto/rand 保证安全性）
func RandomIntLimited(min int, max int) (int, error) {
	if min >= max {
		return 0, fmt.Errorf("min must be less than max")
	}
	rangeSize := big.NewInt(int64(max - min))
	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}
