// Package Captcha 编排「申请验证码并发送短信」的业务流程：
// 生成验证码写入 Redis → 调用短信网关 → 处理发送冷却与失败释放。
//
// handler 只负责取参数和把返回的错误映射为响应，避免同一套逻辑在两个
// handler（公开短信接口、登录态修改密码短信接口）里各写一份。
package Captcha

import (
	"errors"
	"fmt"
	"time"

	"miaoverse/service/security/sms/codemanager"
	"miaoverse/service/security/sms/smsbao"
)

// Expire 验证码有效期：写入 Redis 的过期时间与短信正文中告知用户的有效期共用该值。
const Expire = 5 * time.Minute

// ErrProvider 短信网关发送失败。
// 单独定义哨兵错误，让 handler 能区分「网关失败」与「Redis 等系统异常」，
// 从而在不暴露网关内部信息的前提下给出正确的提示文案。
var ErrProvider = errors.New("sms provider failed")

// Send 为指定业务场景申请验证码并发送短信，成功返回验证码 UUID（即客户端的 code_uuid）。
//
// 返回错误：
//   - codemanager.ErrTooFrequent：同一「场景 + 区号 + 手机号」处于发送冷却期内（调用方应回 429）；
//   - ErrProvider：短信网关发送失败（调用方应回通用文案，不回显网关错误）；
//   - 其他：Redis 等系统异常。
func Send(
	codeManager *codemanager.CodeManager,
	smsSender *smsbao.SmsBaoServant,
	action string,
	region string,
	phone string,
	usage string,
) (string, error) {
	err, code, codeUUID := codeManager.PrepareCodeForPhone(action, region, phone)
	if err != nil {
		return "", err
	}

	if err := smsSender.SendPhoneCaptcha(phone, code, Expire, usage); err != nil {
		// 网关失败时释放冷却，让用户可以立即重试，而不是白等一个冷却周期
		codeManager.ReleaseCooldown(action, region, phone)
		return "", fmt.Errorf("%w: %v", ErrProvider, err)
	}

	return codeUUID, nil
}
