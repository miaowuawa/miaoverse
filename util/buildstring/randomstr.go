package buildstring

import (
	"crypto/rand"
	"errors"
	"math/big"
	"miaoverse/consts"
	"strings"
)

func generateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("随机字符串长度必须大于0")
	}

	charsetLen := big.NewInt(int64(len(consts.Charset)))
	var sb strings.Builder
	sb.Grow(length) // 预分配内存，提升性能

	for i := 0; i < length; i++ {
		// 生成安全的随机索引
		randIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		sb.WriteByte(consts.Charset[randIndex.Int64()])
	}

	return sb.String(), nil
}
