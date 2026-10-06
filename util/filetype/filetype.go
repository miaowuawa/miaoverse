package filetype

import (
	"mime"
	"net/http"
	"strings"

	"miaoverse/consts"
)

// Normalize 将用户传入的分类或 MIME 类型归类为文件大类（uint8 常量，见 consts.FileType*）
func Normalize(value string, mimeType string) uint8 {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "image":
		return consts.FileTypeImage
	case "video":
		return consts.FileTypeVideo
	case "audio":
		return consts.FileTypeAudio
	case "document":
		return consts.FileTypeDocument
	case "other":
		return consts.FileTypeOther
	}

	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(mimeType))
	}
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return consts.FileTypeImage
	case strings.HasPrefix(mediaType, "video/"):
		return consts.FileTypeVideo
	case strings.HasPrefix(mediaType, "audio/"):
		return consts.FileTypeAudio
	case mediaType == "application/pdf", strings.HasPrefix(mediaType, "text/"), strings.Contains(mediaType, "document"):
		return consts.FileTypeDocument
	default:
		return consts.FileTypeOther
	}
}

// safeImageMIMEs 用户图片（贴纸等）允许的安全栅格图片格式白名单。
// 明确排除 SVG 等可内嵌脚本/事件处理器的格式，防止「上传图片藏 JS」造成存储型 XSS。
var safeImageMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// SafeImageMIME 判断 MIME 是否为安全栅格图片格式（jpg/png/gif/webp）。
func SafeImageMIME(mimeType string) bool {
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(mimeType))
	}
	return safeImageMIMEs[mediaType]
}

// DetectSafeImageMIME 嗅探文件头魔数识别真实图片类型（不解码图片内容），
// 返回识别出的 MIME；不是安全栅格图片时 ok=false。
//
// 安全策略：
//   - 使用 net/http.DetectContentType 按魔数识别，SVG/HTML/JS 等会被识别为 text/*、text/xml 或 text/html，不在白名单内直接拒绝；
//   - 声明的 MIME（如 multipart Content-Type）非空时也必须在白名单内，拒绝「声明 image/png 实际为 SVG」的伪装上传；
//   - 返回的是嗅探到的真实类型，调用方应以它作为存储 Content-Type，避免浏览器按伪造类型解析。
func DetectSafeImageMIME(head []byte, declaredMIME string) (string, bool) {
	detected := http.DetectContentType(head)
	if !safeImageMIMEs[detected] {
		return "", false
	}
	if declared := strings.TrimSpace(declaredMIME); declared != "" && !SafeImageMIME(declared) {
		return "", false
	}
	return detected, true
}
