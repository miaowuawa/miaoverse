package codemanager

import (
	"errors"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8"
	"miaoverse/consts"
)

// VerifySceneCode 按「业务场景 + 区号 + 手机号」校验验证码。
// 与 PrepareCodeForPhone 使用同一个 CodeKeyHash，保证发送/校验两侧规则一致；
// 校验通过后验证码立即销毁（一次性），失败不消耗验证码以便用户重试。
func (c *CodeManager) VerifySceneCode(action string, region string, phone string, codeUUID string, inputCode string) (bool, error) {
	return c.VerifyCodeByRegionPhoneMD5(CodeKeyHash(action, region, phone), codeUUID, inputCode)
}

// VerifyCodeByRegionPhoneMD5 验证验证码是否正确
// 参数说明：
//
//	regionPhoneMD5: 「业务场景 + 区号 + 手机号」的摘要（见 CodeKeyHash）
//	codeUUID: 生成验证码时返回的UUID
//	inputCode: 用户输入的验证码（字符串格式，兼容前端传参）
//
// 返回值：
//
//	bool: 验证是否通过
//	error: 错误信息（Redis操作失败/其他系统错误），验证不通过时返回 (false, nil)
func (c *CodeManager) VerifyCodeByRegionPhoneMD5(regionPhoneMD5 string, codeUUID string, inputCode string) (bool, error) {
	codeID := codeUUID + "-" + regionPhoneMD5

	// 1. 读取验证码。key 不存在（验证码过期 / 已被使用 / 场景不匹配 / UUID 错误）
	//    属于「校验不通过」而不是系统异常，其中场景不匹配还可能是跨场景重放尝试，
	//    统一返回 (false, nil)，由调用方按业务错误提示用户。
	storeCodeStr, err := c.Redis.Get(c.Context, codeID).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// 2. 解析 Redis 中的验证码与用户输入的验证码
	storeCode, err := strconv.Atoi(storeCodeStr)
	if err != nil {
		return false, errors.New("验证码格式错误: " + err.Error())
	}
	inputCodeInt, err := strconv.Atoi(inputCode)
	if err != nil {
		return false, errors.New("输入的验证码格式错误: " + err.Error())
	}

	// 3. 不匹配：累加失败次数，达到上限直接作废验证码
	if storeCode != inputCodeInt {
		c.registerFailedAttempt(codeID)
		return false, nil
	}

	// 4. 匹配：一次性销毁验证码与失败计数器。
	//    删除失败按校验失败处理，宁可让用户重新获取，也不让验证码被重复使用。
	pipe := c.Redis.TxPipeline()
	pipe.Del(c.Context, codeID)
	pipe.Del(c.Context, attemptsKey(codeID))
	if _, err := pipe.Exec(c.Context); err != nil {
		log.Printf("删除验证码失败，codeID=%s, err=%v", codeID, err)
		return false, errors.New("验证码验证成功，但删除失败: " + err.Error())
	}

	return true, nil
}

// attemptsKey 返回某个验证码的失败次数计数器 key。
func attemptsKey(codeID string) string {
	return codeID + ":attempts"
}

// registerFailedAttempt 累加同一验证码的校验失败次数。
// 达到 consts.SMSMaxVerifyAttempts 后连同验证码一起删除，强制攻击者重新申请短信：
// 单个验证码因此只有有限次猜测机会，4 位数字验证码无法被在线暴力枚举。
// 计数器 TTL 跟随验证码剩余有效期，验证码过期后计数器自然失效。
func (c *CodeManager) registerFailedAttempt(codeID string) {
	key := attemptsKey(codeID)

	attempts, err := c.Redis.Incr(c.Context, key).Result()
	if err != nil {
		// 计数失败不影响「校验不通过」这一结论，只记日志，避免把客户端错误升级成 500
		log.Printf("记录验证码失败次数失败，codeID=%s, err=%v", codeID, err)
		return
	}

	if attempts == 1 {
		if ttl, err := c.Redis.TTL(c.Context, codeID).Result(); err == nil && ttl > 0 {
			c.Redis.Expire(c.Context, key, ttl)
		}
	}

	if attempts >= int64(consts.SMSMaxVerifyAttempts) {
		c.Redis.Del(c.Context, codeID, key)
	}
}
