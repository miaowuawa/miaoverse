package article

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"miaoverse/consts"
	modelarticle "miaoverse/model/dao/article"
)

// UpdateMetadata 更新文章元数据（只查 MySQL）。
// 只允许更新白名单内的业务字段，mongo_id、id、created_at 等关键字段一律不允许修改。
func (d *ArticleDAO) UpdateMetadata(id uint64, updates map[string]any) error {
	if id == 0 {
		return ErrArticleInvalidID
	}
	if len(updates) == 0 {
		return nil
	}
	filtered := make(map[string]any, len(updates))
	for key, value := range updates {
		switch key {
		case "title", "description", "preview_head", "cover", "status":
			filtered[key] = value
		}
	}
	if len(filtered) == 0 {
		return errors.New("no allowed fields to update")
	}
	if err := d.DB.Model(&modelarticle.Metadata{}).Where("id = ?", id).Updates(filtered).Error; err != nil {
		return fmt.Errorf("update article metadata: %w", err)
	}
	return nil
}

// UpdateArticleStatus 更新文章状态（跨库）：
// MySQL 更新状态并维护 deleted_at，MongoDB 同步状态字段。
func (d *ArticleDAO) UpdateArticleStatus(ctx context.Context, id uint64, status uint8) error {
	if id == 0 {
		return ErrArticleInvalidID
	}
	meta, err := d.QueryMetadataByID(id)
	if err != nil {
		return err
	}

	updates := map[string]any{"status": status}
	if status == consts.ArticleStatusDeleted {
		updates["deleted_at"] = time.Now()
	}
	if err := d.DB.Model(&modelarticle.Metadata{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	objID, err := bson.ObjectIDFromHex(meta.MongoID)
	if err != nil {
		return fmt.Errorf("invalid article mongo_id: %w", err)
	}
	if _, err := d.Mongo.Collection(modelarticle.Article{}.CollectionName()).
		UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": status}}); err != nil {
		return fmt.Errorf("update article body status: %w", err)
	}
	return nil
}

// UpdateArticleBody 更新文章正文（MongoDB），并刷新两库 updated_at。
// 只改正文内容，不动元数据其他字段。
func (d *ArticleDAO) UpdateArticleBody(ctx context.Context, id uint64, content string) error {
	if id == 0 {
		return ErrArticleInvalidID
	}
	meta, err := d.QueryMetadataByID(id)
	if err != nil {
		return err
	}

	objID, err := bson.ObjectIDFromHex(meta.MongoID)
	if err != nil {
		return fmt.Errorf("invalid article mongo_id: %w", err)
	}

	now := time.Now()
	result, err := d.Mongo.Collection(modelarticle.Article{}.CollectionName()).
		UpdateOne(ctx, bson.M{"_id": objID},
			bson.M{"$set": bson.M{"content": content, "updated_at": now}})
	if err != nil {
		return fmt.Errorf("update article body: %w", err)
	}
	if result.MatchedCount == 0 {
		return ErrArticleBodyMissing
	}

	return d.DB.Model(&modelarticle.Metadata{}).Where("id = ?", id).
		Update("updated_at", now).Error
}

// DeleteArticle 软删除文章（跨库）：MySQL 置删除状态并写 deleted_at，MongoDB 同步状态。
func (d *ArticleDAO) DeleteArticle(ctx context.Context, id uint64) error {
	return d.UpdateArticleStatus(ctx, id, consts.ArticleStatusDeleted)
}
