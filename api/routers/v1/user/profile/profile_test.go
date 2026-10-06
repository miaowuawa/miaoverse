package profile

import (
	"testing"

	"miaoverse/consts"
	"miaoverse/model/dto/user/updatereq"
	"miaoverse/service/UserProfile"
)

func strPtr(value string) *string { return &value }
func u8Ptr(value uint8) *uint8    { return &value }

// TestFullFieldsRejectsInvalidValues 覆盖 PUT /user/info 的整体校验：
// 任一字段非法都应整体拒绝，不允许部分落库。
func TestFullFieldsRejectsInvalidValues(t *testing.T) {
	valid := func() *updatereq.ProfileFull {
		return &updatereq.ProfileFull{Username: "miaoverse_ab12", Nickname: "喵呜", Bio: "你好", Gender: 1}
	}

	if _, ok := fullFields(valid()); !ok {
		t.Fatal("valid payload should pass validation")
	}

	cases := []struct {
		name   string
		mutate func(*updatereq.ProfileFull)
	}{
		{"空账号名", func(r *updatereq.ProfileFull) { r.Username = "" }},
		{"账号名带空格", func(r *updatereq.ProfileFull) { r.Username = "ab cd" }},
		{"账号名含中文", func(r *updatereq.ProfileFull) { r.Username = "喵星人" }},
		{"账号名含尖括号", func(r *updatereq.ProfileFull) { r.Username = "<script>" }},
		{"昵称为空", func(r *updatereq.ProfileFull) { r.Nickname = "   " }},
		{"昵称含控制字符", func(r *updatereq.ProfileFull) { r.Nickname = "喵\u200d呜" }},
		{"签名含 NUL", func(r *updatereq.ProfileFull) { r.Bio = "a\x00b" }},
		{"性别越界", func(r *updatereq.ProfileFull) { r.Gender = consts.MaxGenderValue + 1 }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := valid()
			c.mutate(req)
			if fields, ok := fullFields(req); ok {
				t.Fatalf("fullFields should reject %s, got %v", c.name, fields)
			}
		})
	}
}

// TestFullFieldsKeepsIdentityColumnsOut 覆盖越权修复：
// 手机号/区号属于账号身份凭据，不支持更改，因此资料更新接口不会把这两列写入 updates。
// 头像虽然由 handler 单独追加（需要数据库校验文件归属），但必须携带 PermAvatar 权限位。
func TestFullFieldsKeepsIdentityColumnsOut(t *testing.T) {
	req := &updatereq.ProfileFull{Username: "miaoverse_ab12", Nickname: "喵呜", Bio: "", Gender: 0}
	fields, ok := fullFields(req)
	if !ok {
		t.Fatal("valid payload should pass validation")
	}

	columns := map[string]bool{}
	for _, field := range fields {
		columns[field.Column] = true
	}
	for _, forbidden := range []string{"region", "phone", "status", "password"} {
		if columns[forbidden] {
			t.Fatalf("profile update must not write column %q", forbidden)
		}
	}
}

// TestFullFieldsAcceptsAvatarPointer 覆盖「允许修改头像」：
// ProfileFull.Avatar 是可选指针，不传表示不改头像（不会被当成空字符串清空）。
func TestFullFieldsAcceptsAvatarPointer(t *testing.T) {
	withoutAvatar := &updatereq.ProfileFull{Username: "miaoverse_ab12", Nickname: "喵呜"}
	fields, ok := fullFields(withoutAvatar)
	if !ok {
		t.Fatal("payload without avatar should pass validation")
	}
	for _, field := range fields {
		if field.Column == "avatar" {
			t.Fatal("omitted avatar must not be written by fullFields")
		}
	}

	// 与 Region/Avatar 字段在 DTO 中的存在性保持一致（编译期检查 + 语义说明）
	if withoutAvatar.Avatar != nil {
		t.Fatal("avatar should default to nil when omitted")
	}
}

// TestPatchFieldsOnlyIncludesProvidedValues 覆盖 PATCH 语义：未传字段不参与更新。
func TestPatchFieldsOnlyIncludesProvidedValues(t *testing.T) {
	fields, ok := patchFields(&updatereq.ProfilePatch{Nickname: strPtr("新昵称")})
	if !ok {
		t.Fatal("patch with one field should be accepted")
	}
	if len(fields) != 1 || fields[0].Column != "nickname" || fields[0].Value != "新昵称" {
		t.Fatalf("patchFields = %v, want only nickname", fields)
	}
	if fields[0].Perm != UserProfile.PermNickname {
		t.Fatalf("nickname perm = %d, want %d", fields[0].Perm, UserProfile.PermNickname)
	}
}

// patchFields 返回的 ok 只表示「已传字段合法」；空请求体由 handler 在合并头像字段后拒绝。
func TestPatchFieldsEmptyBodyIsValidButEmpty(t *testing.T) {
	fields, ok := patchFields(&updatereq.ProfilePatch{})
	if !ok {
		t.Fatal("empty patch body has no invalid field")
	}
	if len(fields) != 0 {
		t.Fatalf("empty patch body should produce no fields, got %v", fields)
	}
}

// TestPatchFieldsAcceptsAvatarOnlyBody 覆盖「只改头像」的场景：
// avatar 由 handler 单独校验并追加，patchFields 本身不处理它。
func TestPatchFieldsAcceptsAvatarOnlyBody(t *testing.T) {
	fields, ok := patchFields(&updatereq.ProfilePatch{Avatar: strPtr("15b3d25d-66cc-4ddc-9949-33c9e84d8c5d")})
	if !ok {
		t.Fatal("avatar-only patch should not be treated as invalid by patchFields")
	}
	if len(fields) != 0 {
		t.Fatalf("patchFields must not handle avatar itself, got %v", fields)
	}
}

func TestPatchFieldsRejectsInvalidGender(t *testing.T) {
	if _, ok := patchFields(&updatereq.ProfilePatch{Gender: u8Ptr(consts.MaxGenderValue + 1)}); ok {
		t.Fatal("out-of-range gender should be rejected")
	}
}

func TestPatchFieldsAllowsClearingBio(t *testing.T) {
	fields, ok := patchFields(&updatereq.ProfilePatch{Bio: strPtr("")})
	if !ok || len(fields) != 1 || fields[0].Value != "" {
		t.Fatalf("clearing bio should be allowed, got %v (ok=%v)", fields, ok)
	}
}

func TestIsConflict(t *testing.T) {
	if !isConflict(errString("Error 1062: Duplicate entry 'miaoverse' for key 'uk_user_username'")) {
		t.Fatal("MySQL duplicate entry should be treated as conflict")
	}
	if isConflict(errString("connection refused")) {
		t.Fatal("connection error must not be treated as conflict")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
