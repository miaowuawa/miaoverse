package article

import (
	"errors"
	"testing"

	"miaoverse/consts"
	modelarticle "miaoverse/model/dao/article"
)

func TestValidateNovelMeta(t *testing.T) {
	tests := []struct {
		name string
		meta *modelarticle.Metadata
		want error
	}{
		{
			name: "normal article without novel fields",
			meta: &modelarticle.Metadata{Type: consts.ArticleTypeNormal},
			want: nil,
		},
		{
			name: "normal article with novel fields rejected",
			meta: &modelarticle.Metadata{Type: consts.ArticleTypeNormal, NovelID: 1, ChapterID: 2},
			want: ErrArticleInvalidNovel,
		},
		{
			name: "novel root without chapter",
			meta: &modelarticle.Metadata{Type: consts.ArticleTypeNovel},
			want: nil,
		},
		{
			name: "novel root with chapter rejected",
			meta: &modelarticle.Metadata{Type: consts.ArticleTypeNovel, ChapterID: 1},
			want: ErrArticleInvalidNovel,
		},
		{
			name: "novel chapter valid",
			meta: &modelarticle.Metadata{Type: consts.ArticleTypeNovel, NovelID: 5, ChapterID: 3},
			want: nil,
		},
		{
			name: "novel chapter with zero chapter id rejected",
			meta: &modelarticle.Metadata{Type: consts.ArticleTypeNovel, NovelID: 5},
			want: ErrArticleInvalidNovel,
		},
		{
			name: "novel chapter with over limit chapter id rejected",
			meta: &modelarticle.Metadata{Type: consts.ArticleTypeNovel, NovelID: 5, ChapterID: consts.MaxArticleChapterID + 1},
			want: ErrArticleInvalidNovel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNovelMeta(tt.meta)
			if !errors.Is(err, tt.want) {
				t.Fatalf("validateNovelMeta() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestInteractCountFieldAllowed(t *testing.T) {
	for _, field := range []string{"like_count", "comment_count", "share_count", "view_count", "click_count", "repost_count"} {
		if !interactCountFieldAllowed(field) {
			t.Fatalf("interactCountFieldAllowed(%q) = false, want true", field)
		}
	}
	for _, field := range []string{"article_id", "mongo_id", "id", "updated_at; DROP TABLE users; --"} {
		if interactCountFieldAllowed(field) {
			t.Fatalf("interactCountFieldAllowed(%q) = true, want false", field)
		}
	}
}
