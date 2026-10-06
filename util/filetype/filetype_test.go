package filetype

import "testing"

func TestSafeImageMIME(t *testing.T) {
	cases := []struct {
		mime string
		want bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"image/svg+xml", false},
		{"text/html", false},
		{"application/javascript", false},
		{"", false},
		{"IMAGE/PNG", true},
		{"image/png; charset=utf-8", true},
	}
	for _, c := range cases {
		if got := SafeImageMIME(c.mime); got != c.want {
			t.Fatalf("SafeImageMIME(%q) = %v, want %v", c.mime, got, c.want)
		}
	}
}

func TestDetectSafeImageMIME(t *testing.T) {
	pngHead := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}
	jpegHead := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0, 0, 0}
	gifHead := []byte("GIF89a....")
	webpHead := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
	svgHead := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)
	htmlHead := []byte(`<!DOCTYPE html><script>alert(1)</script>`)

	if mime, ok := DetectSafeImageMIME(pngHead, "image/png"); !ok || mime != "image/png" {
		t.Fatalf("png detect = %q/%v, want image/png/true", mime, ok)
	}
	if mime, ok := DetectSafeImageMIME(jpegHead, ""); !ok || mime != "image/jpeg" {
		t.Fatalf("jpeg detect = %q/%v, want image/jpeg/true", mime, ok)
	}
	if mime, ok := DetectSafeImageMIME(gifHead, "image/gif"); !ok || mime != "image/gif" {
		t.Fatalf("gif detect = %q/%v, want image/gif/true", mime, ok)
	}
	if mime, ok := DetectSafeImageMIME(webpHead, "image/webp"); !ok || mime != "image/webp" {
		t.Fatalf("webp detect = %q/%v, want image/webp/true", mime, ok)
	}
	// SVG/HTML 内嵌脚本必须拒绝
	if _, ok := DetectSafeImageMIME(svgHead, "image/svg+xml"); ok {
		t.Fatal("svg must be rejected")
	}
	if _, ok := DetectSafeImageMIME(svgHead, "image/png"); ok {
		t.Fatal("svg declared as png must be rejected")
	}
	if _, ok := DetectSafeImageMIME(htmlHead, "image/png"); ok {
		t.Fatal("html must be rejected")
	}
	// 真实内容是 png 但声明类型不在白名单，同样拒绝
	if _, ok := DetectSafeImageMIME(pngHead, "image/svg+xml"); ok {
		t.Fatal("png declared as svg must be rejected")
	}
}
