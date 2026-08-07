package filetype

import (
	"mime"
	"strings"

	modeluser "miaoverse/model/dao/user"
)

// Normalize 将用户传入的分类或 MIME 类型归类为文件大类（uint8 常量，见 model/dao/user.FileType*）
func Normalize(value string, mimeType string) uint8 {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "image":
		return modeluser.FileTypeImage
	case "video":
		return modeluser.FileTypeVideo
	case "audio":
		return modeluser.FileTypeAudio
	case "document":
		return modeluser.FileTypeDocument
	case "other":
		return modeluser.FileTypeOther
	}

	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(mimeType))
	}
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return modeluser.FileTypeImage
	case strings.HasPrefix(mediaType, "video/"):
		return modeluser.FileTypeVideo
	case strings.HasPrefix(mediaType, "audio/"):
		return modeluser.FileTypeAudio
	case mediaType == "application/pdf", strings.HasPrefix(mediaType, "text/"), strings.Contains(mediaType, "document"):
		return modeluser.FileTypeDocument
	default:
		return modeluser.FileTypeOther
	}
}
