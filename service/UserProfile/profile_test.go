package UserProfile

import (
	"errors"
	"testing"
	"time"

	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
)

type fakeStore struct {
	mask       uint32
	maskErr    error
	updateErr  error
	gotUpdates map[string]any
	gotUserID  uint32
	updateCall int
}

func (f *fakeStore) QueryActivePunishmentMask(userID uint32, now time.Time) (uint32, error) {
	if f.maskErr != nil {
		return 0, f.maskErr
	}
	return f.mask, nil
}

func (f *fakeStore) UpdateProfile(userID uint32, updates map[string]any) (*modeluser.User, error) {
	f.updateCall++
	f.gotUserID = userID
	f.gotUpdates = updates
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return &modeluser.User{ID: userID}, nil
}

func TestNormalizeUsername(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{"trimmed", "  miaoverse_ab12  ", "miaoverse_ab12", true},
		{"dots and dashes", "a.b-c_1", "a.b-c_1", true},
		{"too short", "a", "", false},
		{"leading dot", ".abc", "", false},
		{"space inside", "ab cd", "", false},
		{"chinese", "喵星人", "", false},
		{"angle brackets", "<script>", "", false},
		{"zero width joiner", "ab\u200dcd", "", false},
		{"empty", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := NormalizeUsername(c.input)
			if ok != c.ok || got != c.want {
				t.Fatalf("NormalizeUsername(%q) = (%q, %v), want (%q, %v)", c.input, got, ok, c.want, c.ok)
			}
		})
	}
}

func TestNormalizeUsernameRejectsOverlongValue(t *testing.T) {
	long := make([]byte, 0, consts.MaxUsernameLen+1)
	for i := 0; i <= consts.MaxUsernameLen; i++ {
		long = append(long, 'a')
	}
	if _, ok := NormalizeUsername(string(long)); ok {
		t.Fatalf("NormalizeUsername should reject %d-character username", len(long))
	}
}

func TestNormalizeNicknameCountsRunesNotBytes(t *testing.T) {
	// 30 个汉字 = 90 字节，但只有 30 个字符，必须放行（旧实现按字节长度会误判超长）
	nickname := ""
	for i := 0; i < 30; i++ {
		nickname += "喵"
	}
	got, ok := NormalizeNickname(nickname)
	if !ok || got != nickname {
		t.Fatalf("NormalizeNickname(30 chinese chars) = (%q, %v), want passthrough", got, ok)
	}

	over := nickname + nickname + nickname // 90 字符，超过 64
	if _, ok := NormalizeNickname(over); ok {
		t.Fatal("NormalizeNickname should reject 90-character nickname")
	}
}

func TestNormalizeNicknameRejectsInvisibleChars(t *testing.T) {
	cases := []string{
		"nick\nname",
		"nick\tname",
		"nick\u202ename", // 双向覆盖
		"nick\u200bname", // 零宽空格
		"",
		"   ",
	}
	for _, value := range cases {
		if _, ok := NormalizeNickname(value); ok {
			t.Fatalf("NormalizeNickname(%q) should be rejected", value)
		}
	}
}

func TestNormalizeBio(t *testing.T) {
	got, ok := NormalizeBio("  第一行\r\n第二行\r第三行  ")
	if !ok {
		t.Fatal("NormalizeBio should accept multi-line bio")
	}
	if got != "第一行\n第二行\n第三行" {
		t.Fatalf("NormalizeBio = %q, want normalized newlines and trimmed", got)
	}

	if _, ok := NormalizeBio("bad\x00bio"); ok {
		t.Fatal("NormalizeBio should reject NUL byte")
	}

	over := ""
	for i := 0; i <= consts.MaxBioLen; i++ {
		over += "a"
	}
	if _, ok := NormalizeBio(over); ok {
		t.Fatal("NormalizeBio should reject overlong bio")
	}
}

func TestNormalizeGender(t *testing.T) {
	for value := uint8(0); value <= consts.MaxGenderValue; value++ {
		if _, ok := NormalizeGender(value); !ok {
			t.Fatalf("NormalizeGender(%d) should be accepted", value)
		}
	}
	if _, ok := NormalizeGender(consts.MaxGenderValue + 1); ok {
		t.Fatal("NormalizeGender should reject out-of-range value")
	}
}

func TestUpdateRejectsEmptyFields(t *testing.T) {
	store := &fakeStore{}
	if _, err := Update(store, 10001, nil, time.Now()); !errors.Is(err, ErrNoFields) {
		t.Fatalf("Update(empty) error = %v, want ErrNoFields", err)
	}
	if store.updateCall != 0 {
		t.Fatal("Update should not touch the store when there are no fields")
	}
}

func TestUpdateRejectsPunishedField(t *testing.T) {
	store := &fakeStore{mask: consts.PermSignature}
	fields := []Field{{Column: "bio", Value: "hello", Perm: PermBio}}
	if _, err := Update(store, 10001, fields, time.Now()); !errors.Is(err, ErrPunished) {
		t.Fatalf("Update error = %v, want ErrPunished", err)
	}
	if store.updateCall != 0 {
		t.Fatal("Update must not write when a punished permission bit is involved")
	}
}

func TestUpdateAllowsFieldWithoutPunishment(t *testing.T) {
	// 生效中的惩罚位与本次修改字段无关时（PermComment），修改昵称仍然放行
	store := &fakeStore{mask: consts.PermComment}
	fields := []Field{{Column: "nickname", Value: "新昵称", Perm: PermNickname}}
	if _, err := Update(store, 10001, fields, time.Now()); err != nil {
		t.Fatalf("Update error = %v, want nil", err)
	}
	if store.gotUpdates["nickname"] != "新昵称" {
		t.Fatalf("Update updates = %v, want nickname written", store.gotUpdates)
	}
	if store.gotUserID != 10001 {
		t.Fatalf("Update userID = %d, want 10001", store.gotUserID)
	}
}

func TestUpdateSkipsPunishmentQueryForUnrestrictedFields(t *testing.T) {
	// gender 无需权限位：即使惩罚查询会报错也不应触发查询
	store := &fakeStore{maskErr: errors.New("db down")}
	fields := []Field{{Column: "gender", Value: uint8(1)}}
	if _, err := Update(store, 10001, fields, time.Now()); err != nil {
		t.Fatalf("Update error = %v, want nil", err)
	}
}
