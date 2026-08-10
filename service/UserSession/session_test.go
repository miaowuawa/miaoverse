package UserSession

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/session"
	"miaoverse/consts"
)

// newTestApp 构造带 session 中间件的测试应用：
// 第一个路由写入会话值，第二个路由用新协程上下文读取，模拟同一 cookie 下的两次请求。
func newTestApp(t *testing.T) *fiber.App {
	t.Helper()

	app := fiber.New()
	app.Use(session.New(session.Config{
		Extractor: extractors.FromCookie("mwu_sess_id"),
	}))

	app.Post("/set", func(c fiber.Ctx) error {
		sess := session.FromContext(c)
		sess.Set(consts.SessionPhone, "13800138000")
		sess.Set(consts.SessionRegion, 86)
		sess.Set(consts.SessionUID, 10001)
		return c.JSON(fiber.Map{"ok": true})
	})
	app.Post("/switch", func(c fiber.Ctx) error {
		return SwitchAccount(c, "13800138000", 86, 20002)
	})
	app.Get("/read", func(c fiber.Ctx) error {
		phone, region, ok := CurrentPhoneRegion(c)
		if !ok {
			return c.JSON(fiber.Map{"phone": "", "region": 0, "uid": 0})
		}
		uid, _ := CurrentUID(c)
		return c.JSON(fiber.Map{"phone": phone, "region": region, "uid": uid})
	})

	return app
}

func TestCurrentPhoneRegionAfterLogin(t *testing.T) {
	app := newTestApp(t)

	req := doTestRequest(t, app, newTestRequest(t, "/set", "POST"))
	cookie := req.Cookies()[0]

	resp := doTestRequest(t, app, newTestRequest(t, "/read", "GET", withCookie(cookie)))

	var got struct {
		Phone  string `json:"phone"`
		Region uint16 `json:"region"`
		UID    uint32 `json:"uid"`
	}
	if err := decodeJSON(resp, &got); err != nil {
		t.Fatal(err)
	}
	if got.Phone != "13800138000" || got.Region != 86 || got.UID != 10001 {
		t.Fatalf("session values = %+v, want phone=13800138000 region=86 uid=10001", got)
	}
}

func TestSwitchAccountUpdatesSession(t *testing.T) {
	app := newTestApp(t)

	req := doTestRequest(t, app, newTestRequest(t, "/set", "POST"))
	cookie := req.Cookies()[0]

	// SwitchAccount 会重新生成 session ID，响应中的 Set-Cookie 才是新的会话
	swResp := doTestRequest(t, app, newTestRequest(t, "/switch", "POST", withCookie(cookie)))
	swCookie := swResp.Cookies()[0]

	resp := doTestRequest(t, app, newTestRequest(t, "/read", "GET", withCookie(swCookie)))

	var got struct {
		Phone  string `json:"phone"`
		Region uint16 `json:"region"`
		UID    uint32 `json:"uid"`
	}
	if err := decodeJSON(resp, &got); err != nil {
		t.Fatal(err)
	}
	if got.UID != 20002 {
		t.Fatalf("uid after switch = %d, want 20002", got.UID)
	}
	if got.Phone != "13800138000" || got.Region != 86 {
		t.Fatalf("phone/region after switch = %s/%d, want 13800138000/86", got.Phone, got.Region)
	}
}

func TestCurrentPhoneRegionMissing(t *testing.T) {
	app := fiber.New()
	app.Use(session.New(session.Config{}))
	app.Get("/read", func(c fiber.Ctx) error {
		phone, region, ok := CurrentPhoneRegion(c)
		return c.JSON(fiber.Map{"phone": phone, "region": region, "ok": ok})
	})

	resp := doTestRequest(t, app, newTestRequest(t, "/read", "GET"))

	var got struct {
		Phone  string `json:"phone"`
		Region uint16 `json:"region"`
		OK     bool   `json:"ok"`
	}
	if err := decodeJSON(resp, &got); err != nil {
		t.Fatal(err)
	}
	if got.OK || got.Phone != "" || got.Region != 0 {
		t.Fatalf("CurrentPhoneRegion on empty session = ok:%v phone:%q region:%d, want all empty", got.OK, got.Phone, got.Region)
	}
}
