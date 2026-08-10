package UserSession

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// testRequestOptions 用于链式附加 cookie 等请求配置。
type testRequestOptions func(*http.Request)

func withCookie(cookie *http.Cookie) testRequestOptions {
	return func(req *http.Request) {
		req.AddCookie(cookie)
	}
}

func newTestRequest(t *testing.T, path, method string, opts ...testRequestOptions) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, path, bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("build request %s %s: %v", method, path, err)
	}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

// doTestRequest 发送测试请求，断言无错误后返回响应。
func doTestRequest(t *testing.T, app *fiber.App, req *http.Request) *http.Response {
	t.Helper()

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", req.Method, req.URL, err)
	}
	return resp
}

func decodeJSON(resp *http.Response, out any) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}
