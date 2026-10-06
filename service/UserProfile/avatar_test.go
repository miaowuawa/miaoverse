package UserProfile

import (
	"errors"
	"testing"

	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
)

const testAvatarUUID = "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d"

type fakeAvatarStore struct {
	files map[string]modeluser.File
	err   error
}

func (f *fakeAvatarStore) QueryActiveFilesByUUIDsBatch(fileUUIDs []string) (map[string]modeluser.File, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.files, nil
}

func publicImageFile(userID uint32) modeluser.File {
	return modeluser.File{
		UUID:       testAvatarUUID,
		UserID:     userID,
		FileType:   consts.FileTypeImage,
		Permission: consts.FilePermissionPublic,
		Status:     consts.FileStatusActive,
	}
}

func TestNormalizeAvatarFileAcceptsOwnPublicImage(t *testing.T) {
	store := &fakeAvatarStore{files: map[string]modeluser.File{testAvatarUUID: publicImageFile(10001)}}

	got, err := NormalizeAvatarFile(store, 10001, "  "+testAvatarUUID+"  ")
	if err != nil {
		t.Fatalf("NormalizeAvatarFile error = %v", err)
	}
	if got != testAvatarUUID {
		t.Fatalf("NormalizeAvatarFile = %q, want %q", got, testAvatarUUID)
	}
}

func TestNormalizeAvatarFileRejects(t *testing.T) {
	own := publicImageFile(10001)

	private := own
	private.Permission = consts.FilePermissionNone

	notImage := own
	notImage.FileType = consts.FileTypeDocument

	otherUser := own
	otherUser.UserID = 20002

	cases := []struct {
		name   string
		files  map[string]modeluser.File
		raw    string
		expect error
	}{
		{"空 UUID", map[string]modeluser.File{testAvatarUUID: own}, "   ", ErrAvatarEmpty},
		{"文件不存在", map[string]modeluser.File{}, testAvatarUUID, ErrAvatarInvalid},
		{"非公开文件", map[string]modeluser.File{testAvatarUUID: private}, testAvatarUUID, ErrAvatarInvalid},
		{"非图片文件", map[string]modeluser.File{testAvatarUUID: notImage}, testAvatarUUID, ErrAvatarInvalid},
		{"他人文件", map[string]modeluser.File{testAvatarUUID: otherUser}, testAvatarUUID, ErrAvatarInvalid},
		{"超长 UUID", map[string]modeluser.File{testAvatarUUID: own}, string(make([]byte, consts.MaxAvatarLen+1)), ErrAvatarInvalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := &fakeAvatarStore{files: c.files}
			if _, err := NormalizeAvatarFile(store, 10001, c.raw); !errors.Is(err, c.expect) {
				t.Fatalf("NormalizeAvatarFile error = %v, want %v", err, c.expect)
			}
		})
	}
}

func TestAvatarFieldCarriesPermAvatar(t *testing.T) {
	store := &fakeAvatarStore{files: map[string]modeluser.File{testAvatarUUID: publicImageFile(10001)}}

	field, err := AvatarField(store, 10001, testAvatarUUID)
	if err != nil {
		t.Fatalf("AvatarField error = %v", err)
	}
	if field.Column != "avatar" {
		t.Fatalf("AvatarField column = %q, want avatar", field.Column)
	}
	if field.Perm != consts.PermAvatar {
		t.Fatalf("AvatarField perm = %d, want %d", field.Perm, consts.PermAvatar)
	}
	if field.Value != testAvatarUUID {
		t.Fatalf("AvatarField value = %v, want %q", field.Value, testAvatarUUID)
	}
}

func TestAvatarFieldPropagatesStoreError(t *testing.T) {
	store := &fakeAvatarStore{err: errors.New("db down")}
	if _, err := AvatarField(store, 10001, testAvatarUUID); err == nil {
		t.Fatal("AvatarField should propagate store error")
	}
}
