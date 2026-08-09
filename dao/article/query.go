package article

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
	modelarticle "miaoverse/model/dao/article"
)

// QueryMetadataByID 只查询文章元数据（MySQL），不触达 MongoDB。
func (d *ArticleDAO) QueryMetadataByID(id uint64) (*modelarticle.Metadata, error) {
	if id == 0 {
		return nil, ErrArticleInvalidID
	}
	var meta modelarticle.Metadata
	err := d.DB.Where("id = ?", id).First(&meta).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

// QueryArticleByID 跨库查询文章详情：元数据（MySQL）+ 正文（MongoDB）。
// 没有"只按 id 拉正文不拉元数据"的查询，正文查询必须经元数据定位 Mongo 文档。
func (d *ArticleDAO) QueryArticleByID(ctx context.Context, id uint64) (*modelarticle.Detail, error) {
	meta, err := d.QueryMetadataByID(id)
	if err != nil {
		return nil, err
	}

	body, err := d.queryBodyByMongoID(ctx, meta.MongoID)
	if err != nil {
		return nil, err
	}

	return &modelarticle.Detail{Metadata: *meta, Article: *body}, nil
}

// QueryChapterByNovel 跨库查询小说根文章下的指定章节：先查 metadata（novel_id + chapter_id），再查正文。
// 通过 chapter_id 获取"文章 id（novel_id）下的第几章节"。
func (d *ArticleDAO) QueryChapterByNovel(ctx context.Context, novelID uint64, chapterID uint64) (*modelarticle.Detail, error) {
	if novelID == 0 || chapterID == 0 {
		return nil, ErrArticleInvalidID
	}
	var meta modelarticle.Metadata
	err := d.DB.Where("novel_id = ? AND chapter_id = ?", novelID, chapterID).First(&meta).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}

	body, err := d.queryBodyByMongoID(ctx, meta.MongoID)
	if err != nil {
		return nil, err
	}
	return &modelarticle.Detail{Metadata: meta, Article: *body}, nil
}

// QueryChaptersByNovel 分页查询小说根文章下的章节元数据列表（只查 MySQL），按章节号升序。
func (d *ArticleDAO) QueryChaptersByNovel(novelID uint64, offset, limit int) ([]modelarticle.Metadata, error) {
	if novelID == 0 {
		return nil, ErrArticleInvalidID
	}
	var list []modelarticle.Metadata
	err := d.DB.Where("novel_id = ?", novelID).
		Order("chapter_id ASC").
		Offset(offset).Limit(limit).
		Find(&list).Error
	return list, err
}

// QueryMetadatasByUser 分页查询某用户可见状态下的文章元数据列表（只查 MySQL）。
// status 传 consts.ArticleStatusNormal 等具体状态；如不过滤传 255（全量）。
// 默认只返回非章节文章（novel_id = 0），章节通过 QueryChaptersByNovel 获取。
func (d *ArticleDAO) QueryMetadatasByUser(userID uint32, status uint8, offset, limit int) ([]modelarticle.Metadata, error) {
	var list []modelarticle.Metadata
	q := d.DB.Where("user_id = ? AND novel_id = 0", userID)
	if status != 255 {
		q = q.Where("status = ?", status)
	}
	err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}

// CountMetadatasByUser 统计某用户文章元数据数量（不含章节）。
func (d *ArticleDAO) CountMetadatasByUser(userID uint32, status uint8) (int64, error) {
	var count int64
	q := d.DB.Model(&modelarticle.Metadata{}).Where("user_id = ? AND novel_id = 0", userID)
	if status != 255 {
		q = q.Where("status = ?", status)
	}
	err := q.Count(&count).Error
	return count, err
}

// QueryInteractCount 查询单篇文章互动计数。
func (d *ArticleDAO) QueryInteractCount(articleID uint64) (*modelarticle.ArticleInteractCount, error) {
	if articleID == 0 {
		return nil, ErrArticleInvalidID
	}
	var count modelarticle.ArticleInteractCount
	err := d.DB.Where("article_id = ?", articleID).First(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// QueryInteractCounts 批量查询文章互动计数，返回 article_id → 计数 map。
func (d *ArticleDAO) QueryInteractCounts(articleIDs []uint64) (map[uint64]modelarticle.ArticleInteractCount, error) {
	result := map[uint64]modelarticle.ArticleInteractCount{}
	if len(articleIDs) == 0 {
		return result, nil
	}
	var list []modelarticle.ArticleInteractCount
	if err := d.DB.Where("article_id IN ?", articleIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, c := range list {
		result[c.ArticleID] = c
	}
	return result, nil
}

// IncrementInteractCount 原子增减文章互动计数。
// 只允许白名单字段，防止任意字段名拼入 SQL。
func (d *ArticleDAO) IncrementInteractCount(articleID uint64, field string, delta int64) error {
	if articleID == 0 {
		return ErrArticleInvalidID
	}
	if !interactCountFieldAllowed(field) {
		return fmt.Errorf("invalid interact count field: %s", field)
	}
	return d.DB.Model(&modelarticle.ArticleInteractCount{}).
		Where("article_id = ?", articleID).
		UpdateColumn(field, gorm.Expr(field+" + ?", delta)).Error
}

// interactCountFieldAllowed 互动计数字段白名单
func interactCountFieldAllowed(field string) bool {
	switch field {
	case "like_count", "comment_count", "share_count", "view_count", "click_count", "repost_count":
		return true
	}
	return false
}

// QueryMetadatasByIDs 按 id 批量查询文章元数据（只查 MySQL）。
// 用于互动/评论等场景回查文章归属，避免逐条跨库。
func (d *ArticleDAO) QueryMetadatasByIDs(ids []uint64) ([]modelarticle.Metadata, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []modelarticle.Metadata
	err := d.DB.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

// queryBodyByMongoID 按 MongoDB _id 查询文章正文。
// 仅内部使用：正文查询一律以元数据为入口，保证返回内容必带元数据。
func (d *ArticleDAO) queryBodyByMongoID(ctx context.Context, mongoID string) (*modelarticle.Article, error) {
	objID, err := bson.ObjectIDFromHex(mongoID)
	if err != nil {
		return nil, fmt.Errorf("invalid article mongo_id: %w", err)
	}

	var body modelarticle.Article
	err = d.Mongo.Collection(modelarticle.Article{}.CollectionName()).
		FindOne(ctx, bson.M{"_id": objID}).Decode(&body)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrArticleBodyMissing
	}
	if err != nil {
		return nil, err
	}
	return &body, nil
}
