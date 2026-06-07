package filetype

import (
	"mime"
	"strings"
)

func Normalize(value string, mimeType string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "image", "video", "audio", "document", "other":
		return value
	}

	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(mimeType))
	}
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return "image"
	case strings.HasPrefix(mediaType, "video/"):
		return "video"
	case strings.HasPrefix(mediaType, "audio/"):
		return "audio"
	case mediaType == "application/pdf", strings.HasPrefix(mediaType, "text/"), strings.Contains(mediaType, "document"):
		return "document"
	default:
		return "other"
	}
}
