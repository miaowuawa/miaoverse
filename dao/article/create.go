package article

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"gorm.io/gorm"
	"miaoverse/consts"
	modelarticle "miaoverse/model/dao/article"
)

var (
	// ErrArticleNotFound 文章元数据不存在
	ErrArticleNotFound = errors.New("article metadata not found")
	// ErrArticleBodyMissing 元数据存在但 MongoDB 正文缺失（数据不一致）
	ErrArticleBodyMissing = errors.New("article body missing in mongodb")
	// ErrArticleInvalidID 非法的文章 id
	ErrArticleInvalidID = errors.New("invalid article id")
	// ErrArticleInvalidNovel 小说章节元数据不合法（chapter_id/novel_id 缺失或非法）
	ErrArticleInvalidNovel = errors.New("invalid novel chapter metadata")
)

// CreateArticle 创建文章：先写 MySQL 元数据与互动计数，再写 MongoDB 正文。
// 两库无分布式事务：Mongo 写入失败时补偿删除 MySQL 行，避免留下孤儿元数据。
// 小说章节：meta.Type 为 ArticleTypeNovel 时，meta.NovelID 必须指向小说根文章、meta.ChapterID 必须 >= 1；
// 根文章（非章节）NovelID 为 0。
func (d *ArticleDAO) CreateArticle(ctx context.Context, meta *modelarticle.Metadata, body *modelarticle.Article) error {
	if meta == nil || body == nil {
		return errors.New("article metadata and body are required")
	}
	if meta.UserID == 0 {
		return ErrArticleInvalidID
	}
	if err := validateNovelMeta(meta); err != nil {
		return err
	}
	// 章节文章：确认指向的小说根文章真实存在且为 novel 类型，防止孤儿章节
	if meta.NovelID != 0 {
		root, err := d.QueryMetadataByID(meta.NovelID)
		if err != nil {
			return fmt.Errorf("query novel root article: %w", err)
		}
		if root.Type != consts.ArticleTypeNovel || root.NovelID != 0 {
			return ErrArticleInvalidNovel
		}
	}

	// 1. 预生成 MongoDB _id，写入 MySQL 元数据时即带上关联（插入后才回填会留下 mongo_id 为空的行）
	objID := bson.NewObjectID()
	meta.MongoID = objID.Hex()

	// 2. MySQL 写入元数据 + 互动计数行（事务），拿自增 id（对外暴露的文章 id）
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(meta).Error; err != nil {
			return err
		}
		// 同步创建互动计数行，保证后续计数自增有目标行
		return tx.Create(&modelarticle.ArticleInteractCount{ArticleID: meta.ID}).Error
	})
	if err != nil {
		return fmt.Errorf("create article metadata: %w", err)
	}

	// 3. MongoDB 写入正文，用 MySQL 自增 id 作跨库关联键
	now := time.Now()
	body.ID = meta.ID
	body.ObjectID = objID
	body.Type = meta.Type
	if body.CreatedAt.IsZero() {
		body.CreatedAt = now
	}
	if body.UpdatedAt.IsZero() {
		body.UpdatedAt = now
	}
	if _, err := d.Mongo.Collection(modelarticle.Article{}.CollectionName()).InsertOne(ctx, body); err != nil {
		// 补偿：正文写入失败则删除 MySQL 元数据（互动计数行随外键级联删除），保证两库一致
		if rmErr := d.DB.Unscoped().Delete(&modelarticle.Metadata{}, meta.ID).Error; rmErr != nil {
			return fmt.Errorf("create article body failed (%v) and rollback metadata failed: %v", err, rmErr)
		}
		return fmt.Errorf("create article body: %w", err)
	}

	return nil
}

// validateNovelMeta 校验小说章节元数据合法性。
// 规则：novel 类型且 NovelID > 0（章节）时必须带 1..MaxArticleChapterID 的 chapter_id；
// 非章节文章（普通/转载/小说根）NovelID 与 ChapterID 必须为 0。
func validateNovelMeta(meta *modelarticle.Metadata) error {
	if meta.Type != consts.ArticleTypeNovel {
		if meta.NovelID != 0 || meta.ChapterID != 0 {
			return ErrArticleInvalidNovel
		}
		return nil
	}
	if meta.NovelID == 0 {
		// 小说根文章：非章节，不允许带 chapter_id
		if meta.ChapterID != 0 {
			return ErrArticleInvalidNovel
		}
		return nil
	}
	if meta.ChapterID < 1 || meta.ChapterID > consts.MaxArticleChapterID {
		return ErrArticleInvalidNovel
	}
	return nil
}
