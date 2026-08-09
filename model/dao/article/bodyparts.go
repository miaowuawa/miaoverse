package article

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"miaoverse/consts"
)

// Article 文章正文（MongoDB）。
// ID 与 Metadata.ID 一致，作为跨库关联键；ObjectID 为 MongoDB 内部 _id，不对外暴露。
type Article struct {
	ObjectID           bson.ObjectID `bson:"_id,omitempty" json:"-"`
	ID                 uint64        `bson:"article_id" json:"id"`
	ReferenceArticleID uint64        `bson:"reference_article_id" json:"reference_id"`
	Type               uint8         `bson:"type" json:"type"`
	Content            string        `bson:"content" json:"content"`
	Status             uint8         `bson:"status" json:"status"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func (Article) CollectionName() string {
	return consts.ArticleMongoCollection
}
