package conf

import "miaoverse/consts"

func (c *AppConfig) UploadMaxFileSizeBytes() int64 {
	if c == nil || c.Upload.MaxFileSizeBytes <= 0 {
		return consts.DefaultUploadMaxFileSizeBytes
	}
	return c.Upload.MaxFileSizeBytes
}
