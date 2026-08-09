package dao

import (
	"miaoverse/dao/article"
	"miaoverse/dao/content"
	"miaoverse/dao/interacts"
	"miaoverse/dao/user"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

// NewUserDao 创建用户DAO实例
func NewUserDao(db *gorm.DB) *user.UserDAO {
	return &user.UserDAO{DB: db}
}

// NewContentDao 创建内容域DAO实例
func NewContentDao(db *gorm.DB) *content.ContentDAO {
	return &content.ContentDAO{DB: db}
}

// NewInteractsDao 创建互动域DAO实例
func NewInteractsDao(db *gorm.DB) *interacts.InteractsDAO {
	return &interacts.InteractsDAO{DB: db}
}

// NewArticleDao 创建文章域DAO实例（跨 MySQL + MongoDB）
func NewArticleDao(db *gorm.DB, mongoDB *mongo.Database) *article.ArticleDAO {
	return &article.ArticleDAO{DB: db, Mongo: mongoDB}
}
