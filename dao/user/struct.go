package user

import (
	"gorm.io/gorm"
)

// UserDAO 用户DAO结构体，持有DB连接，挂载用户相关方法
type UserDAO struct {
	DB *gorm.DB // 关联根DB，复用连接池
}
