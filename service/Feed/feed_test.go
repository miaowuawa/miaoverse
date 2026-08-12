package Feed

import (
	"testing"
	"time"

	"miaoverse/consts"
	modelarticle "miaoverse/model/dao/article"
	modelmoment "miaoverse/model/dao/moment"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
)

func TestMomentToItem(t *testing.T) {
	now := time.Now()
	m := &modelmoment.Moment{
		ID:         1,
		UserID:     10001,
		Title:      "标题",
		Content:    "内容",
		Status:     consts.MomentStatusNormal,
		Permission: consts.MomentPermissionPublic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	author := modeluser.User{ID: 10001, Nickname: "喵"}
	meta := modelmoment.MomentInteractCount{MomentID: 1, LikeCount: 3, CommentCount: 2, ShareCount: 1}

	item := momentToItem(m, author, meta, true)
	if item.Type != consts.ContentTypeMoment {
		t.Fatalf("type = %q, want %q", item.Type, consts.ContentTypeMoment)
	}
	if item.ID != 1 || item.UserID != 10001 || item.Content != "内容" {
		t.Fatalf("basic fields mismatch: %+v", item)
	}
	if item.Stats.Likes != 3 || item.Stats.Comments != 2 || item.Stats.Shares != 1 {
		t.Fatalf("stats mismatch: %+v", item.Stats)
	}
	if !item.IsLiked || !item.Full {
		t.Fatalf("is_liked=%v full=%v, want true/true", item.IsLiked, item.Full)
	}
	if item.Author.Nickname != "喵" {
		t.Fatalf("author mismatch: %+v", item.Author)
	}
}

func TestArticleToItem(t *testing.T) {
	now := time.Now()
	a := &modelarticle.Metadata{
		ID:          2,
		UserID:      10002,
		Title:       "文章标题",
		Description: "摘要",
		Cover:       "cover.jpg",
		Type:        consts.ArticleTypeNormal,
		Status:      consts.ArticleStatusNormal,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	author := modeluser.User{ID: 10002, Nickname: "爱音"}
	meta := modelarticle.ArticleInteractCount{ArticleID: 2, LikeCount: 5, ViewCount: 100}

	item := articleToItem(a, author, meta, false)
	if item.Type != consts.ContentTypeArticle {
		t.Fatalf("type = %q, want %q", item.Type, consts.ContentTypeArticle)
	}
	if item.Title != "文章标题" || item.Description != "摘要" || item.Cover != "cover.jpg" {
		t.Fatalf("article fields mismatch: %+v", item)
	}
	if item.Stats.Likes != 5 || item.Stats.Comments != 0 {
		t.Fatalf("stats mismatch: %+v", item.Stats)
	}
	if item.IsLiked || item.Full {
		t.Fatalf("is_liked=%v full=%v, want false/false", item.IsLiked, item.Full)
	}
}

func TestAssembleSortsByCreatedAtDesc(t *testing.T) {
	base := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	moments := []modelmoment.Moment{
		{ID: 1, UserID: 1, CreatedAt: base.Add(-2 * time.Hour)},
		{ID: 2, UserID: 1, CreatedAt: base},
	}
	articles := []modelarticle.Metadata{
		{ID: 3, UserID: 1, CreatedAt: base.Add(-1 * time.Hour)},
	}

	items := mergeSorted(moments, articles)
	if len(items) != 3 {
		t.Fatalf("len = %d, want 3", len(items))
	}
	wantOrder := []uint64{2, 3, 1}
	for i, want := range wantOrder {
		if items[i].ID != want {
			t.Fatalf("items[%d].ID = %d, want %d", i, items[i].ID, want)
		}
	}
}

func TestMergeSortedStableOnEqualTime(t *testing.T) {
	base := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	moments := []modelmoment.Moment{
		{ID: 1, UserID: 1, CreatedAt: base},
	}
	articles := []modelarticle.Metadata{
		{ID: 2, UserID: 1, CreatedAt: base},
	}

	items := mergeSorted(moments, articles)
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	// 同时间按 id 倒序：2 在前
	if items[0].ID != 2 || items[1].ID != 1 {
		t.Fatalf("order = [%d, %d], want [2, 1]", items[0].ID, items[1].ID)
	}
}

// mergeSorted 将动态与文章条目按发布时间倒序合并（同时间按 id 倒序）。
// 与 assemble 的排序口径一致，抽出来便于单测。
func mergeSorted(moments []modelmoment.Moment, articles []modelarticle.Metadata) []resp.FeedItem {
	items := make([]resp.FeedItem, 0, len(moments)+len(articles))
	for i := range moments {
		items = append(items, momentToItem(&moments[i], modeluser.User{}, modelmoment.MomentInteractCount{}, false))
	}
	for i := range articles {
		items = append(items, articleToItem(&articles[i], modeluser.User{}, modelarticle.ArticleInteractCount{}, false))
	}
	sortFeedItems(items)
	return items
}
