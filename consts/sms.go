package consts

import "time"

// 短信验证码域的频次限制常量。
//
// 背景：验证码为 4 位数字（1000-9999），如果不限制校验次数，攻击者可以在验证码
// 有效期内在线枚举；如果不同时限制发送频次，攻击者又能通过「反复申请新验证码」
// 无限续接枚举预算。两者必须一起限制。

const (
	// SMSSendCooldown 同一「业务场景 + 区号 + 手机号」两次申请验证码的最小间隔。
	SMSSendCooldown = 60 * time.Second
	// SMSMaxVerifyAttempts 单个验证码允许的最大校验失败次数，达到上限后验证码立即失效，
	// 用户必须重新申请（每个验证码因此最多只有 5 次猜测机会）。
	SMSMaxVerifyAttempts = 5
)
