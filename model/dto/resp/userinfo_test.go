package resp

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	modeluser "miaoverse/model/dao/user"
)

// phoneEchoApp 构造一个应用：路由内返回带 phone 字段的响应，用于断言 omitempty 行为。
func phoneEchoApp(info UserInfo) *fiber.App {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(CodeWithMsgUserInfo{Code: 200, Msg: "ok", User: info})
	})
	return app
}

func decodeUserInfo(t *testing.T, info UserInfo) map[string]any {
	t.Helper()
	app := phoneEchoApp(info)
	res, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	user, ok := body["user"].(map[string]any)
	if !ok {
		t.Fatalf("response has no user object: %v", body)
	}
	return user
}

// TestUserInfoPhoneOmittedForOtherUsers 覆盖隐私要求：
// 查询他人资料时 phone 字段必须完全不出现在响应里（而不是返回空串）。
func TestUserInfoPhoneOmittedForOtherUsers(t *testing.T) {
	user := decodeUserInfo(t, UserInfo{User: modeluser.User{ID: 20002, Username: "other"}})
	if _, exists := user["phone"]; exists {
		t.Fatalf("phone must be omitted when querying other users, got %v", user["phone"])
	}
}

// TestUserInfoPhoneReturnedForSelf 覆盖本人查询：返回打码后的手机号。
func TestUserInfoPhoneReturnedForSelf(t *testing.T) {
	user := decodeUserInfo(t, UserInfo{
		User:  modeluser.User{ID: 10001, Username: "self"},
		Phone: "+86 138****8000",
	})
	got, ok := user["phone"].(string)
	if !ok || got != "+86 138****8000" {
		t.Fatalf("phone = %v, want masked phone for self query", user["phone"])
	}
}

// TestCodeWithMsgUserPhoneOmitted 覆盖 /user/me 等接口：未绑定手机号时不输出 phone。
func TestCodeWithMsgUserPhoneOmitted(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(CodeWithMsgUser{
			Code: 200,
			Msg:  "ok",
			User: modeluser.User{ID: 10001, Username: "self"},
		})
	})

	res, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, exists := body["phone"]; exists {
		t.Fatalf("phone must be omitted when the session has no bound phone, got %v", body["phone"])
	}
}
