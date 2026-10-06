package conf

import "miaoverse/consts"

func (c *AppConfig) UploadMaxFileSizeBytes() int64 {
	if c == nil || c.Upload.MaxFileSizeBytes <= 0 {
		return consts.DefaultUploadMaxFileSizeBytes
	}
	return c.Upload.MaxFileSizeBytes
}

// UploadMaxStickerSizeBytes 单张贴纸最大字节数（默认 10MB）。
func (c *AppConfig) UploadMaxStickerSizeBytes() int64 {
	if c == nil || c.Upload.MaxStickerSizeBytes <= 0 {
		return consts.DefaultStickerMaxFileSizeBytes
	}
	return c.Upload.MaxStickerSizeBytes
}
