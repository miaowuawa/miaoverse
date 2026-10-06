package codemanager

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/utils/v2"
	"miaoverse/consts"
	"miaoverse/util"
)

// ErrTooFrequent 同一「场景 + 手机号」在冷却期内重复申请验证码。
var ErrTooFrequent = errors.New("sms code requested too frequently")

func generateCode() int {
	code, _ := util.Maths.RandomIntLimited(1000, 10000)
	return code
}

// CodeKeyHash 计算「业务场景 + 区号 + 手机号」的摘要，作为验证码存储 key 的固定组成部分。
// 发送与校验必须走同一个函数，避免两侧拼接规则漂移导致验证码永远校验不通过。
// 场景参与摘要后，各业务场景的验证码互不通用（登录验证码不能用于修改密码）。
func CodeKeyHash(action string, region string, phone string) string {
	return util.MD5Hash.HashStr(action + ":" + region + ":" + phone)
}

// CooldownKey 返回发送冷却 key。
func CooldownKey(action string, region string, phone string) string {
	return "sms:cooldown:" + CodeKeyHash(action, region, phone)
}

// PrepareCodeForPhone 为指定业务场景生成验证码并写入 Redis，返回明文验证码与本次验证码 UUID。
// action 必须是 consts 中定义的业务场景，由调用方（handler）完成白名单校验。
//
// 同一「场景 + 区号 + 手机号」在 consts.SMSSendCooldown 内只允许申请一次，重复申请返回
// ErrTooFrequent（调用方应回 429），用于防止短信轰炸与「反复申请验证码 + 暴力枚举」的组合攻击。
func (c *CodeManager) PrepareCodeForPhone(action string, region string, phone string) (error, string, string) {
	// SetNX 保证并发请求下只有一个能拿到发送权，不会出现同时发出多条短信
	acquired, err := c.Redis.SetNX(c.Context, CooldownKey(action, region, phone), 1, consts.SMSSendCooldown).Result()
	if err != nil {
		return err, "", ""
	}
	if !acquired {
		return ErrTooFrequent, "", ""
	}

	codeUUID := utils.UUIDv4()
	codeID := codeUUID + "-" + CodeKeyHash(action, region, phone)
	code := generateCode()

	if err := c.Redis.Set(c.Context, codeID, code, c.CodeExpireTime).Err(); err != nil {
		// 写码失败时释放冷却，避免用户被无谓地锁在冷却期内无法重试
		c.ReleaseCooldown(action, region, phone)
		return err, "", ""
	}
	return nil, strconv.Itoa(code), codeUUID
}

// ReleaseCooldown 释放发送冷却。短信网关发送失败时由调用方调用，
// 让用户能立即重试，而不是白等一个冷却周期。
func (c *CodeManager) ReleaseCooldown(action string, region string, phone string) {
	if err := c.Redis.Del(c.Context, CooldownKey(action, region, phone)).Err(); err != nil {
		log.Printf("释放短信发送冷却失败，err=%v", err)
	}
}
