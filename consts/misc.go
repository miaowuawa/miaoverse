package consts

import "time"

// 杂项常量：时间格式、定时任务参数、上传默认值、UA 类型。

const TimeFormat = "2006-01-02 15:04:05"

const (
	// MetaSyncReconcileWindow 增量校准窗口：只校准最近更新过的动态
	MetaSyncReconcileWindow = 30 * time.Minute
	// MetaSyncBatchSize 每批处理的动态数量，避免单次 SQL 过大
	MetaSyncBatchSize = 500
)

const DefaultUploadMaxFileSizeBytes int64 = 20 * 1024 * 1024

const (
	UATypeUnknown string = "unknown" // 未知类型
	UATypePC             = "pc"      // PC端
	UATypeWAP            = "wap"     // 移动端/WAP
	UATypeBot            = "bot"     // 爬虫/Bot
)
