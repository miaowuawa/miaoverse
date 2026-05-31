package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMessageByLangDefaultsToZhCN(t *testing.T) {
	resetCatalog()

	got := MessageByLang("", OKLogin)
	if got != "登录成功" {
		t.Fatalf("MessageByLang empty lang = %q, want %q", got, "登录成功")
	}
}

func TestMessageByLangNormalizesAcceptLanguage(t *testing.T) {
	resetCatalog()

	got := MessageByLang("zh-CN,zh;q=0.9,en;q=0.8", ErrBadRequest)
	if got != "请求错误，请检查参数" {
		t.Fatalf("MessageByLang zh-CN header = %q, want %q", got, "请求错误，请检查参数")
	}
}

func TestMessageByLangFallbacksToKey(t *testing.T) {
	resetCatalog()

	const key MessageKey = "missing.key"
	got := MessageByLang(LangZhCN, key)
	if got != string(key) {
		t.Fatalf("MessageByLang missing key = %q, want %q", got, string(key))
	}
}

func TestLoadDirLoadsTomlMessagesAndPlurals(t *testing.T) {
	resetCatalog()
	dir := t.TempDir()
	content := []byte(`
language = "en-US"

[messages]
"ok.login" = "Login successful"
"error.sms_provider" = "Provider error: {{error}}"

[plurals."mail.unread"]
zero = "No unread messages"
one = "{{count}} unread message"
other = "{{count}} unread messages"
`)
	if err := os.WriteFile(filepath.Join(dir, "en-US.toml"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadDir(dir); err != nil {
		t.Fatal(err)
	}

	if got := MessageByLang("en-US", OKLogin); got != "Login successful" {
		t.Fatalf("MessageByLang en-US = %q, want %q", got, "Login successful")
	}
	if got := Render("en-US", ErrSMSProvider, Data{"error": "timeout"}); got != "Provider error: timeout" {
		t.Fatalf("Render named data = %q, want %q", got, "Provider error: timeout")
	}
	if got := PluralByLang("en-US", "mail.unread", 0, nil); got != "No unread messages" {
		t.Fatalf("PluralByLang zero = %q, want %q", got, "No unread messages")
	}
	if got := PluralByLang("en-US", "mail.unread", 1, nil); got != "1 unread message" {
		t.Fatalf("PluralByLang one = %q, want %q", got, "1 unread message")
	}
	if got := PluralByLang("en-US", "mail.unread", 3, nil); got != "3 unread messages" {
		t.Fatalf("PluralByLang other = %q, want %q", got, "3 unread messages")
	}
}
