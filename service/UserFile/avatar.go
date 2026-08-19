package UserFile

import (
	"github.com/google/uuid"
	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
)

// ValidateAvatarUUID 校验头像文件 UUID：
//   - 必须是合法 UUID 格式；
//   - 文件必须存在、为 active 状态、属于当前用户、且为图片类型（安全：禁止引用他人/非图片/失效文件）。
//
// 头像文件必须公开（permission=0）才能被所有人查看且不受拉黑/屏蔽影响，由调用方校验。
func ValidateAvatarUUID(userServant interface {
	QueryActiveFilesByUUIDsBatch(fileUUIDs []string) (map[string]modeluser.File, error)
}, uid uint32, avatarUUID string) (*modeluser.File, bool) {
	if _, err := uuid.Parse(avatarUUID); err != nil {
		return nil, false
	}
	files, err := userServant.QueryActiveFilesByUUIDsBatch([]string{avatarUUID})
	if err != nil {
		return nil, false
	}
	file, ok := files[avatarUUID]
	if !ok || file.UserID != uid || file.FileType != consts.FileTypeImage {
		return nil, false
	}
	return &file, true
}
