package filetype

import (
	"mime"
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
