package content

import "gorm.io/gorm"

// ContentDAO 内容域DAO结构体，持有DB连接，挂载动态/互动/评论/通知相关方法
type ContentDAO struct {
	DB *gorm.DB // 关联根DB，复用连接池
}
