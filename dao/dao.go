package dao

import (
	"miaoverse/dao/user"

	"gorm.io/gorm"
)

// NewUserDao 创建用户DAO实例
func NewUserDao(db *gorm.DB) *user.UserDAO {
	return &user.UserDAO{DB: db}
}
