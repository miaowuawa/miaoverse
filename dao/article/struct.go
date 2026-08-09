package article

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

// ArticleDAO 文章域DAO结构体，同时持有 MySQL 与 MongoDB 句柄。
// Metadata 存 MySQL（article_meta），正文存 MongoDB（article），查询需跨数据库。
type ArticleDAO struct {
	DB    *gorm.DB        // MySQL，复用连接池
	Mongo *mongo.Database // MongoDB
}
