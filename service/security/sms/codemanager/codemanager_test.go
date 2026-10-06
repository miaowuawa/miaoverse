package codemanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"miaoverse/consts"
)

func newTestManager(t *testing.T) *CodeManager {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return &CodeManager{
		Redis:          client,
		CodeExpireTime: 5 * time.Minute,
		Context:        context.Background(),
	}
}

func TestPrepareCodeForPhoneEnforcesCooldown(t *testing.T) {
	manager := newTestManager(t)

	if err, _, _ := manager.PrepareCodeForPhone(consts.ActionLogin, "86", "13800138000"); err != nil {
		t.Fatalf("first PrepareCodeForPhone error = %v", err)
	}

	// 冷却期内重复申请被拒绝（防短信轰炸）
	err, _, _ := manager.PrepareCodeForPhone(consts.ActionLogin, "86", "13800138000")
	if !errors.Is(err, ErrTooFrequent) {
		t.Fatalf("second PrepareCodeForPhone error = %v, want ErrTooFrequent", err)
	}

	// 冷却按「场景 + 手机号」隔离：换手机号不受影响
	if err, _, _ := manager.PrepareCodeForPhone(consts.ActionLogin, "86", "13800138001"); err != nil {
		t.Fatalf("other phone PrepareCodeForPhone error = %v", err)
	}
	// 同一手机号换场景同样不受影响
	if err, _, _ := manager.PrepareCodeForPhone(consts.ActionChangePassword, "86", "13800138000"); err != nil {
		t.Fatalf("other action PrepareCodeForPhone error = %v", err)
	}

	// 释放冷却后可以立即重试（短信网关失败时使用）
	manager.ReleaseCooldown(consts.ActionLogin, "86", "13800138000")
	if err, _, _ := manager.PrepareCodeForPhone(consts.ActionLogin, "86", "13800138000"); err != nil {
		t.Fatalf("PrepareCodeForPhone after release error = %v", err)
	}
}

func TestVerifySceneCodeInvalidatesCodeAfterTooManyFailures(t *testing.T) {
	manager := newTestManager(t)

	err, code, codeUUID := manager.PrepareCodeForPhone(consts.ActionChangePassword, "86", "13800138000")
	if err != nil {
		t.Fatalf("PrepareCodeForPhone error = %v", err)
	}

	wrong := "000000"
	if wrong == code {
		wrong = "111111"
	}

	for i := 0; i < consts.SMSMaxVerifyAttempts; i++ {
		ok, err := manager.VerifySceneCode(consts.ActionChangePassword, "86", "13800138000", codeUUID, wrong)
		if err != nil {
			t.Fatalf("attempt %d error = %v", i+1, err)
		}
		if ok {
			t.Fatalf("attempt %d with wrong code must not pass", i+1)
		}
	}

	// 达到失败上限后验证码作废，即使随后输入正确验证码也不再通过，
	// 避免 4 位数字验证码被在线暴力枚举
	ok, err := manager.VerifySceneCode(consts.ActionChangePassword, "86", "13800138000", codeUUID, code)
	if err != nil {
		t.Fatalf("verify after lockout error = %v", err)
	}
	if ok {
		t.Fatal("verification code must be invalidated after too many failed attempts")
	}
}

func TestCodeKeyHashIsolatesActions(t *testing.T) {
	login := CodeKeyHash(consts.ActionLogin, "86", "13800138000")
	change := CodeKeyHash(consts.ActionChangePassword, "86", "13800138000")
	if login == change {
		t.Fatal("code key hash must differ between sms actions")
	}

	other := CodeKeyHash(consts.ActionLogin, "86", "13800138001")
	if login == other {
		t.Fatal("code key hash must differ between phone numbers")
	}

	if got := CodeKeyHash(consts.ActionLogin, "86", "13800138000"); got != login {
		t.Fatal("code key hash must be deterministic")
	}
}

func TestVerifySceneCodeRejectsCrossSceneReplay(t *testing.T) {
	manager := newTestManager(t)

	err, code, codeUUID := manager.PrepareCodeForPhone(consts.ActionLogin, "86", "13800138000")
	if err != nil {
		t.Fatalf("PrepareCodeForPhone error = %v", err)
	}

	// 为登录申请的验证码不能用于修改密码
	ok, err := manager.VerifySceneCode(consts.ActionChangePassword, "86", "13800138000", codeUUID, code)
	if err != nil {
		t.Fatalf("VerifySceneCode error = %v", err)
	}
	if ok {
		t.Fatal("login code must not pass verification for the change-password scene")
	}

	// 登录场景仍然可以正常校验
	ok, err = manager.VerifySceneCode(consts.ActionLogin, "86", "13800138000", codeUUID, code)
	if err != nil {
		t.Fatalf("VerifySceneCode error = %v", err)
	}
	if !ok {
		t.Fatal("login code should pass verification for the login scene")
	}
}

func TestVerifySceneCodeIsSingleUse(t *testing.T) {
	manager := newTestManager(t)

	err, code, codeUUID := manager.PrepareCodeForPhone(consts.ActionChangePassword, "86", "13800138000")
	if err != nil {
		t.Fatalf("PrepareCodeForPhone error = %v", err)
	}

	ok, err := manager.VerifySceneCode(consts.ActionChangePassword, "86", "13800138000", codeUUID, code)
	if err != nil || !ok {
		t.Fatalf("first verify = (%v, %v), want (true, nil)", ok, err)
	}

	// 验证码一次性：通过后立即销毁，不能重放
	ok, err = manager.VerifySceneCode(consts.ActionChangePassword, "86", "13800138000", codeUUID, code)
	if err == nil && ok {
		t.Fatal("verification code must not be reusable")
	}
}

func TestVerifySceneCodeRejectsWrongCodeAndKeepsIt(t *testing.T) {
	manager := newTestManager(t)

	err, code, codeUUID := manager.PrepareCodeForPhone(consts.ActionChangePassword, "86", "13800138000")
	if err != nil {
		t.Fatalf("PrepareCodeForPhone error = %v", err)
	}

	ok, err := manager.VerifySceneCode(consts.ActionChangePassword, "86", "13800138000", codeUUID, "000000")
	if err != nil {
		t.Fatalf("VerifySceneCode error = %v", err)
	}
	if ok {
		t.Fatal("wrong code must not pass verification")
	}

	// 输错不应消耗验证码，用户仍可重试
	ok, err = manager.VerifySceneCode(consts.ActionChangePassword, "86", "13800138000", codeUUID, code)
	if err != nil || !ok {
		t.Fatalf("retry with correct code = (%v, %v), want (true, nil)", ok, err)
	}
}

func TestVerifySceneCodeRejectsUnknownUUID(t *testing.T) {
	manager := newTestManager(t)

	err, code, _ := manager.PrepareCodeForPhone(consts.ActionChangePassword, "86", "13800138000")
	if err != nil {
		t.Fatalf("PrepareCodeForPhone error = %v", err)
	}

	ok, err := manager.VerifySceneCode(consts.ActionChangePassword, "86", "13800138000", "9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a", code)
	if err == nil && ok {
		t.Fatal("unknown code uuid must not pass verification")
	}
}
