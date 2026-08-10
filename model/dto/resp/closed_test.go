package resp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
)

func closedUser() *modeluser.User {
	return &modeluser.User{
		ID:       10001,
		Username: "user_old",
		Nickname: "old_nickname",
		Bio:      "我的旧签名",
		Status:   consts.UserStatusClosed,
	}
}

// maskAndEcho 构造一个应用：路由内对用户调用 MaskClosedAccount 后原样返回，用于断言遮蔽结果。
func maskAndEcho(u *modeluser.User) *fiber.App {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		MaskClosedAccount(c, u)
		return c.JSON(u)
	})
	return app
}

func TestMaskClosedAccountReplacesUsernameAndBio(t *testing.T) {
	app := maskAndEcho(closedUser())

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got modeluser.User
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Username != "已注销的账号" {
		t.Fatalf("username = %q, want %q", got.Username, "已注销的账号")
	}
	if got.Bio != "用户已注销。虽然我不知道 TA 在三次元过得好不好，但我替这个账号感谢你最后的回眸。 —— 站长留" {
		t.Fatalf("bio = %q, want closed account message", got.Bio)
	}
	if got.Nickname != "old_nickname" {
		t.Fatalf("nickname = %q, want unchanged", got.Nickname)
	}
	if got.Status != consts.UserStatusClosed {
		t.Fatalf("status = %d, want unchanged", got.Status)
	}
}

func TestMaskClosedAccountIgnoresOtherStatuses(t *testing.T) {
	for _, status := range []uint8{consts.UserStatusActive, consts.UserStatusBanned, consts.UserStatusDisabled} {
		u := closedUser()
		u.Status = status
		app := maskAndEcho(u)

		resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
		if err != nil {
			t.Fatal(err)
		}
		var got modeluser.User
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()

		if got.Username != "user_old" || got.Bio != "我的旧签名" {
			t.Fatalf("status %d: username/bio should be unchanged, got %q / %q", status, got.Username, got.Bio)
		}
	}
}

func TestMaskClosedAccountNil(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		MaskClosedAccount(c, nil)
		return c.SendStatus(http.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (no panic)", resp.StatusCode)
	}
}
