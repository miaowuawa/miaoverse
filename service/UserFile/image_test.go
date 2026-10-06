package UserFile

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"testing"
)

// buildUploadFileHeader 构造 multipart 上传的 *multipart.FileHeader（模拟真实上传请求）。
func buildUploadFileHeader(t *testing.T, filename string, contentType string, content []byte) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("CreatePart: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	form, err := multipart.NewReader(&buf, writer.Boundary()).ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("ReadForm: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expect 1 file, got %d", len(files))
	}
	return files[0]
}

var (
	pngBytes  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}
	jpegBytes = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0, 0, 0}
	svgBytes  = []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)
	htmlBytes = []byte(`<!DOCTYPE html><script>alert(1)</script>`)
)

func TestDetectSafeUploadedImage(t *testing.T) {
	// 安全栅格图片放行，返回嗅探出的真实 MIME
	if mime, ok, err := DetectSafeUploadedImage(buildUploadFileHeader(t, "a.png", "image/png", pngBytes), "image/png"); err != nil || !ok || mime != "image/png" {
		t.Fatalf("png = %q/%v/%v, want image/png/true/nil", mime, ok, err)
	}
	if mime, ok, err := DetectSafeUploadedImage(buildUploadFileHeader(t, "a.jpg", "image/jpeg", jpegBytes), "image/jpeg"); err != nil || !ok || mime != "image/jpeg" {
		t.Fatalf("jpeg = %q/%v/%v, want image/jpeg/true/nil", mime, ok, err)
	}

	// 「图片藏 JS」：SVG/HTML 一律拒绝（含声明成 png 的伪装上传）
	if _, ok, err := DetectSafeUploadedImage(buildUploadFileHeader(t, "a.svg", "image/svg+xml", svgBytes), "image/svg+xml"); err != nil || ok {
		t.Fatalf("svg = %v/%v, want false/nil", ok, err)
	}
	if _, ok, err := DetectSafeUploadedImage(buildUploadFileHeader(t, "a.png", "image/png", svgBytes), "image/png"); err != nil || ok {
		t.Fatalf("svg declared png = %v/%v, want false/nil", ok, err)
	}
	if _, ok, err := DetectSafeUploadedImage(buildUploadFileHeader(t, "a.png", "image/png", htmlBytes), "image/png"); err != nil || ok {
		t.Fatalf("html declared png = %v/%v, want false/nil", ok, err)
	}

	// 真实内容是 png 但声明类型不在白名单（如 image/svg+xml），同样拒绝
	if _, ok, err := DetectSafeUploadedImage(buildUploadFileHeader(t, "a.png", "image/svg+xml", pngBytes), "image/svg+xml"); err != nil || ok {
		t.Fatalf("png declared svg = %v/%v, want false/nil", ok, err)
	}
}

func TestReadHead(t *testing.T) {
	head, err := ReadHead(buildUploadFileHeader(t, "a.png", "image/png", pngBytes), 512)
	if err != nil {
		t.Fatalf("ReadHead: %v", err)
	}
	if !bytes.Equal(head, pngBytes) {
		t.Fatalf("ReadHead = %v, want %v", head, pngBytes)
	}

	// 小于嗅探长度的文件按实际长度返回（不报错）
	short := []byte{0x89, 'P'}
	head, err = ReadHead(buildUploadFileHeader(t, "b.png", "image/png", short), 512)
	if err != nil {
		t.Fatalf("ReadHead short: %v", err)
	}
	if !bytes.Equal(head, short) {
		t.Fatalf("ReadHead short = %v, want %v", head, short)
	}
}
