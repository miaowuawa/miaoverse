package UserFile

import (
	"testing"

	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
)

type fakeFileDAO struct {
	files map[string]modeluser.File
}

func (f *fakeFileDAO) QueryActiveFilesByUUIDsBatch(fileUUIDs []string) (map[string]modeluser.File, error) {
	result := map[string]modeluser.File{}
	for _, uuid := range fileUUIDs {
		if file, ok := f.files[uuid]; ok {
			result[uuid] = file
		}
	}
	return result, nil
}

func TestValidateAvatarUUID(t *testing.T) {
	const uid = uint32(10001)
	validUUID := "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d"
	otherUUID := "25b3d25d-66cc-4ddc-9949-33c9e84d8c5d"

	dao := &fakeFileDAO{files: map[string]modeluser.File{
		validUUID: {UUID: validUUID, UserID: uid, FileType: consts.FileTypeImage, Status: consts.FileStatusActive},
		otherUUID: {UUID: otherUUID, UserID: 20002, FileType: consts.FileTypeImage, Status: consts.FileStatusActive},
	}}

	cases := []struct {
		name       string
		avatarUUID string
		wantOK     bool
	}{
		{name: "valid own image", avatarUUID: validUUID, wantOK: true},
		{name: "invalid uuid format", avatarUUID: "not-a-uuid", wantOK: false},
		{name: "empty uuid", avatarUUID: "", wantOK: false},
		{name: "other user's file", avatarUUID: otherUUID, wantOK: false},
		{name: "nonexistent uuid", avatarUUID: "35b3d25d-66cc-4ddc-9949-33c9e84d8c5d", wantOK: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file, ok := ValidateAvatarUUID(dao, uid, tc.avatarUUID)
			if ok != tc.wantOK {
				t.Fatalf("ValidateAvatarUUID(%q) ok = %v, want %v", tc.avatarUUID, ok, tc.wantOK)
			}
			if ok && file == nil {
				t.Fatalf("ValidateAvatarUUID(%q) returned ok but nil file", tc.avatarUUID)
			}
		})
	}
}
