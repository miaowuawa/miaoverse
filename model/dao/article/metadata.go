package article

import "time"

// Metadata 文章元数据（MySQL）。
// ID 为 MySQL 自增主键，是暴露给前端的文章 id；MongoID 仅用于关联 MongoDB 文档，不对外返回。
// 小说（ArticleTypeNovel）按章节存储：每章一条独立文章记录，NovelID 指向小说根文章 id，
// ChapterID 表示该记录是小说根文章下的第几章；非小说文章两者为 0。
type Metadata struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	MongoID     string     `gorm:"column:mongo_id;type:varchar(24);not null;default:''" json:"-"`
	UserID      uint32     `gorm:"column:user_id;not null;index" json:"user_id"`
	Title       string     `gorm:"column:title;type:varchar(255);not null;default:''" json:"title"`
	Description string     `gorm:"column:description;type:varchar(500);not null;default:''" json:"description"`
	PreviewHead string     `gorm:"column:preview_head;type:varchar(300);not null;default:''" json:"head"` //文章前一小部分
	Cover       string     `gorm:"column:cover;type:varchar(255);not null;default:''" json:"cover"`
	Type        uint8      `gorm:"column:type;not null;default:0" json:"type"`               // 0: 普通 1: 转载 2: 小说（分章节）
	NovelID     uint64     `gorm:"column:novel_id;not null;default:0;index" json:"novel_id"` // 所属小说根文章 id，0 表示非章节
	ChapterID   uint64     `gorm:"column:chapter_id;not null;default:0" json:"chapter_id"`   // 章节号（novel 类型生效），1 起
	Status      uint8      `gorm:"column:status;not null;default:0" json:"status"`           // 0: 正常 1: 删除 2: 草稿 3: 限制传播 4: 屏蔽
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;default:null" json:"deleted_at"`
}

func (Metadata) TableName() string {
	return "article_meta"
}

// Detail 跨库查询的聚合结果：元数据 + 正文。
type Detail struct {
	Metadata Metadata `json:"metadata"`
	Article  Article  `json:"article"`
}
