package conf

const DefaultUploadMaxFileSizeBytes int64 = 20 * 1024 * 1024

func (c *AppConfig) UploadMaxFileSizeBytes() int64 {
	if c == nil || c.Upload.MaxFileSizeBytes <= 0 {
		return DefaultUploadMaxFileSizeBytes
	}
	return c.Upload.MaxFileSizeBytes
}
