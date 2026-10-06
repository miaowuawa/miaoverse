package Sticker

import "testing"

const testUUID = "123e4567-e89b-12d3-a456-426614174000"

func TestBuildTokenAndExtractTokens(t *testing.T) {
	token := BuildToken(testUUID)
	if token != "[sticker:"+testUUID+"]" {
		t.Fatalf("BuildToken = %q", token)
	}

	if got := ExtractTokens("你好 [sticker:" + testUUID + "] 世界"); len(got) != 1 || got[0] != testUUID {
		t.Fatalf("ExtractTokens single = %v", got)
	}
	if got := ExtractTokens("纯文本评论"); len(got) != 0 {
		t.Fatalf("ExtractTokens plain = %v", got)
	}
	// 一条评论最多一个贴纸：多个标记会被检测出来（由 CheckCommentSticker 拒绝）
	two := "a [sticker:" + testUUID + "] b [sticker:" + testUUID + "]"
	if got := ExtractTokens(two); len(got) != 2 {
		t.Fatalf("ExtractTokens multi = %v", got)
	}
	// 非法 UUID 形式的方括号内容不误判为贴纸
	if got := ExtractTokens("[sticker:not-a-uuid]"); len(got) != 0 {
		t.Fatalf("ExtractTokens invalid = %v", got)
	}
	// 大小写不敏感，统一转小写
	upper := "[sticker:123E4567-E89B-12D3-A456-426614174000]"
	if got := ExtractTokens(upper); len(got) != 1 || got[0] != testUUID {
		t.Fatalf("ExtractTokens upper = %v", got)
	}
}

func TestNormalizeName(t *testing.T) {
	if got := NormalizeName("  你好  ", 10); got != "你好" {
		t.Fatalf("NormalizeName trim = %q", got)
	}
	// 按 rune 截断，不破坏多字节字符
	if got := NormalizeName("贴纸贴纸贴纸", 4); got != "贴纸贴纸" {
		t.Fatalf("NormalizeName truncate = %q", got)
	}
	if got := NormalizeName("   ", 10); got != "" {
		t.Fatalf("NormalizeName blank = %q", got)
	}
}

func TestValidateUUID(t *testing.T) {
	if v, ok := ValidateUUID(" " + testUUID + " "); !ok || v != testUUID {
		t.Fatalf("ValidateUUID ok = %q/%v", v, ok)
	}
	if _, ok := ValidateUUID("not-a-uuid"); ok {
		t.Fatal("ValidateUUID must reject invalid uuid")
	}
	if _, ok := ValidateUUID(""); ok {
		t.Fatal("ValidateUUID must reject empty")
	}
}
