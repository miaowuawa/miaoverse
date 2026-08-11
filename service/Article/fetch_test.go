package Article

import (
	"strings"
	"testing"

	modelarticle "miaoverse/model/dao/article"
)

func TestTruncatePercent(t *testing.T) {
	if got := truncatePercent("", 60); got != "" {
		t.Fatalf("empty input: got %q", got)
	}
	if got := truncatePercent("你好世界", 100); got != "你好世界" {
		t.Fatalf("100%%: got %q", got)
	}
	if got := truncatePercent("你好世界", 50); got != "你好" {
		t.Fatalf("50%%: got %q", got)
	}
	if got := truncatePercent("abcdef", 60); got != "abc" {
		t.Fatalf("60%%: got %q", got)
	}
	// 百分比截断按 rune 计算，不能截出半个汉字
	if got := truncatePercent("喵喵喵喵", 60); got != "喵喵" {
		t.Fatalf("rune-safe 60%%: got %q", got)
	}
}

func TestSegmentSlice(t *testing.T) {
	size := segmentSize()
	content := strings.Repeat("a", size*2+5)
	if got := segmentSlice(content, 1); len([]rune(got)) != size {
		t.Fatalf("seq 1 length = %d, want %d", len([]rune(got)), size)
	}
	// 中间段：恰好一个分段大小
	if got := segmentSlice(content, 2); len([]rune(got)) != size {
		t.Fatalf("seq 2 length = %d, want %d", len([]rune(got)), size)
	}
	// 末段：余量
	if got := segmentSlice(content, 3); got != "aaaaa" {
		t.Fatalf("last segment: got %q", got)
	}
	if got := segmentSlice("abc", 3); got != "" {
		t.Fatalf("out of range seq: got %q", got)
	}
}

func TestSegmentCount(t *testing.T) {
	size := segmentSize()
	if got := segmentCount(strings.Repeat("a", size)); got != 0 {
		t.Fatalf("at size: got %d, want 0", got)
	}
	if got := segmentCount(strings.Repeat("a", size+1)); got != 2 {
		t.Fatalf("over size: got %d, want 2", got)
	}
	if got := segmentCount(""); got != 0 {
		t.Fatalf("empty: got %d, want 0", got)
	}
}

func TestDeliverableContent(t *testing.T) {
	normal := &modelarticle.Metadata{ChapterID: 0}
	novelCh1 := &modelarticle.Metadata{ChapterID: 1}
	novelCh3 := &modelarticle.Metadata{ChapterID: 3}

	content := strings.Repeat("喵", 100)

	// 登录用户始终拿到完整正文
	if got, full := deliverableContent(content, normal, true); got != content || !full {
		t.Fatalf("logged in normal: got len=%d full=%v", len([]rune(got)), full)
	}
	if got, full := deliverableContent(content, novelCh3, true); got != content || !full {
		t.Fatalf("logged in novel ch3: got len=%d full=%v", len([]rune(got)), full)
	}

	// 未登录普通文章：截断前 60%
	got, full := deliverableContent(content, normal, false)
	if full || got != string([]rune(content)[:60]) {
		t.Fatalf("anonymous normal: got len=%d full=%v", len([]rune(got)), full)
	}

	// 未登录小说前 2 章：完整正文
	if got, full := deliverableContent(content, novelCh1, false); got != content || !full {
		t.Fatalf("anonymous ch1: got len=%d full=%v", len([]rune(got)), full)
	}

	// 未登录第 3 章起：不返回正文
	if got, full := deliverableContent(content, novelCh3, false); got != "" || full {
		t.Fatalf("anonymous ch3: got len=%d full=%v", len([]rune(got)), full)
	}
}

func TestSegmentBoundaries(t *testing.T) {
	size := segmentSize()
	// 恰好一个分段：不需要分段（segmentCount 0）
	if got := segmentCount(strings.Repeat("a", size)); got != 0 {
		t.Fatalf("segmentCount at boundary = %d, want 0", got)
	}
	// 超过一个分段：按 seq 边界切片不越界
	content := strings.Repeat("a", size*3+7)
	last := segmentSlice(content, segmentCount(content))
	if last != strings.Repeat("a", 7) {
		t.Fatalf("last segment mismatch: got len=%d", len([]rune(last)))
	}
}
