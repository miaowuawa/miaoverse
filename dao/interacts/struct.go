package interacts

import "gorm.io/gorm"

// InteractsDAO 互动域DAO结构体，持有DB连接，挂载互动/评论相关方法
type InteractsDAO struct {
	DB *gorm.DB // 关联根DB，复用连接池
}
