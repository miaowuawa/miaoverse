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

	item := momentToItem(m, author, meta, true, []string{"uuid-1"})
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
	if len(item.Images) != 1 || item.Images[0] != "uuid-1" {
		t.Fatalf("images mismatch: %+v", item.Images)
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
		items = append(items, momentToItem(&moments[i], modeluser.User{}, modelmoment.MomentInteractCount{}, false, nil))
	}
	for i := range articles {
		items = append(items, articleToItem(&articles[i], modeluser.User{}, modelarticle.ArticleInteractCount{}, false))
	}
	sortFeedItems(items)
	return items
}

func TestSortHotItems(t *testing.T) {
	base := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	items := []resp.FeedItem{
		{ID: 1, CreatedAt: base},
		{ID: 2, CreatedAt: base, Stats: resp.MomentStats{Likes: 5}},
		{ID: 3, CreatedAt: base.Add(-1 * time.Hour), Stats: resp.MomentStats{Likes: 5}},
	}
	sortHotItems(items)
	// 点赞 5 的在前（同赞按时间倒序：2 先于 3），最后是 0 赞的 1
	if items[0].ID != 2 || items[1].ID != 3 || items[2].ID != 1 {
		t.Fatalf("order = [%d, %d, %d], want [2, 3, 1]", items[0].ID, items[1].ID, items[2].ID)
	}
}

func TestPageSlice(t *testing.T) {
	items := []resp.FeedItem{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}
	cases := []struct {
		name    string
		offset  int
		limit   int
		wantIDs []uint64
	}{
		{name: "first page", offset: 0, limit: 2, wantIDs: []uint64{1, 2}},
		{name: "middle page", offset: 2, limit: 2, wantIDs: []uint64{3, 4}},
		{name: "last partial page", offset: 4, limit: 2, wantIDs: []uint64{5}},
		{name: "offset beyond end", offset: 5, limit: 2, wantIDs: nil},
		{name: "empty source", offset: 0, limit: 2, wantIDs: nil},
	}
	for _, tc := range cases {
		src := items
		if tc.name == "empty source" {
			src = nil
		}
		got := pageSlice(src, tc.offset, tc.limit)
		if len(got) != len(tc.wantIDs) {
			t.Fatalf("%s: len = %d, want %d", tc.name, len(got), len(tc.wantIDs))
		}
		for i, want := range tc.wantIDs {
			if got[i].ID != want {
				t.Fatalf("%s: items[%d].ID = %d, want %d", tc.name, i, got[i].ID, want)
			}
		}
	}
}

// TestUserHotPageWindow 验证用户内容流 hot 分页的取数窗口：
// 第 N 页应取 offset+limit 条后再从合并序列中截取 [offset, offset+limit)。
func TestUserHotPageWindow(t *testing.T) {
	base := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	// 模拟第 2 页（offset=2, limit=2）：hot 模式各分类取 offset+limit=4 条
	offset, limit := 2, 2
	fetchLimit := offset + limit

	// 动态 4 条（模拟 SQL 已按点赞倒序），文章 4 条
	moments := []modelmoment.Moment{
		{ID: 1, CreatedAt: base.Add(-1 * time.Hour)},
		{ID: 2, CreatedAt: base.Add(-2 * time.Hour)},
		{ID: 3, CreatedAt: base.Add(-3 * time.Hour)},
		{ID: 4, CreatedAt: base.Add(-4 * time.Hour)},
	}
	articles := []modelarticle.Metadata{
		{ID: 5, CreatedAt: base},
		{ID: 6, CreatedAt: base.Add(-5 * time.Hour)},
		{ID: 7, CreatedAt: base.Add(-6 * time.Hour)},
		{ID: 8, CreatedAt: base.Add(-7 * time.Hour)},
	}
	if len(moments) < fetchLimit || len(articles) < fetchLimit {
		t.Fatalf("test data too small")
	}
	moments = moments[:fetchLimit]
	articles = articles[:fetchLimit]

	items := make([]resp.FeedItem, 0, len(moments)+len(articles))
	for i := range moments {
		items = append(items, momentToItem(&moments[i], modeluser.User{}, modelmoment.MomentInteractCount{}, false, nil))
	}
	for i := range articles {
		items = append(items, articleToItem(&articles[i], modeluser.User{}, modelarticle.ArticleInteractCount{}, false))
	}
	sortHotItems(items)
	page := pageSlice(items, offset, limit)

	if len(page) != limit {
		t.Fatalf("page len = %d, want %d", len(page), limit)
	}
	// 全局热序（此时全为 0 赞，按时间倒序）应为 [5,1,2,3,4,6,7,8]，第 2 页取 [2,3]
	if page[0].ID != 2 || page[1].ID != 3 {
		t.Fatalf("page IDs = [%d, %d], want [2, 3]", page[0].ID, page[1].ID)
	}
}
