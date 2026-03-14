package sms

import (
	"fmt"
	"math"
	"math/big"
	"time"
)

// 辅助函数：反转字符串（保留前导0）
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ValidateAvalue a计算：平方一下整个时间戳，反转，不去掉开头的0
func ValidateAvalue(a string) (bool, error) {

	// 把整个a值反转过来
	// 1. 反转字符串还原平方值
	squareStr := reverseString(a)

	// 2. 安全解析平方值（用big.Int防止溢出）
	squareBig := new(big.Int)
	_, ok := squareBig.SetString(squareStr, 10)
	if !ok {
		return false, fmt.Errorf("无效的数值格式：%s", a)
	}

	// 3. 计算平方根（大数运算，避免溢出）
	tsBig := new(big.Int).Sqrt(squareBig)
	// 验证平方根的平方是否等于原数（防止非整数平方根）
	verifySquare := new(big.Int).Mul(tsBig, tsBig)
	if verifySquare.Cmp(squareBig) != 0 {
		return false, fmt.Errorf("不是有效的时间戳平方值")
	}

	// 4. 转为int64（时间戳不会超过int64范围）
	tsvalid := tsBig.IsUint64()
	if !tsvalid {
		return false, fmt.Errorf("时间戳异常")
	}
	ts := tsBig.Int64()

	// 5. 验证时间范围（允许±1000ms）
	now := time.Now().UnixMilli()
	// 时间差绝对值 ≤ 1000ms
	if ts < now-1000 || ts > now+1000 {
		return false, fmt.Errorf("时间戳过期，当前时间：%d，请求时间：%d，差值：%dms",
			now, ts, math.Abs(float64(now-ts)))
	}

	return true, nil
}
