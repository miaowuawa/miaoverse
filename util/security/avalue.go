package security

import (
	"fmt"
	"math/big"
	"time"
)

func ValidateAvalue(a string) (bool, error) {
	squareStr := reverseString(a)

	squareBig := new(big.Int)
	_, ok := squareBig.SetString(squareStr, 10)
	if !ok {
		return false, fmt.Errorf("无效的数值格式：%s", a)
	}

	tsBig := new(big.Int).Sqrt(squareBig)
	verifySquare := new(big.Int).Mul(tsBig, tsBig)
	if verifySquare.Cmp(squareBig) != 0 {
		return false, fmt.Errorf("不是有效的时间戳平方值")
	}

	if !tsBig.IsUint64() {
		return false, fmt.Errorf("时间戳异常")
	}
	ts := tsBig.Int64()

	now := time.Now().UnixMilli()
	if ts < now-1000 || ts > now+1000 {
		return false, fmt.Errorf("时间戳过期，当前时间：%d，请求时间：%d，差值：%dms",
			now, ts, absInt64(now-ts))
	}

	return true, nil
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
