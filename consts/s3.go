package consts

import "time"

// S3 域常量：临时签名版本、签名时钟容差、临时链接时长默认值与上限。

const (
	TempSignatureVersion = "v1"
)

const (
	TempSignatureClockGap = time.Minute
)

const (
	DefaultTempSignatureDuration = 10 * time.Minute
	DefaultTempLinkDuration      = 5 * time.Minute
	MaxTempLinkDuration          = 7 * 24 * time.Hour
)
