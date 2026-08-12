package moment

import "time"

// MomentFile 动态-图片关联表：一条动态可挂多张图片，按 sort 排序展示。
// 只存文件 UUID，不存 S3 object key / 原始 URL，避免向客户端暴露存储细节。
type MomentFile struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	MomentID  uint64    `gorm:"not null;index:idx_moment_file_moment_sort" json:"moment_id"`
	FileUUID  string    `gorm:"column:file_uuid;not null;size:36" json:"file_uuid"`
	Sort      uint32    `gorm:"not null;default:0" json:"sort"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (MomentFile) TableName() string {
	return "moment_file"
}
