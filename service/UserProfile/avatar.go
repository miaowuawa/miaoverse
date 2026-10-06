package UserProfile

import (
	"errors"
	"strings"

	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/service/UserFile"
)

var (
	// ErrAvatarEmpty 头像文件 UUID 为空。
	ErrAvatarEmpty = errors.New("empty avatar uuid")
	// ErrAvatarInvalid 头像文件不合法：不存在、非本人、非 active 图片，或不是公开文件。
	ErrAvatarInvalid = errors.New("invalid avatar file")
)

// PermAvatar 修改头像所需的权限位。
const PermAvatar = consts.PermAvatar

// avatarStore 头像文件校验所需的最小 DAO 能力。
type avatarStore interface {
	QueryActiveFilesByUUIDsBatch(fileUUIDs []string) (map[string]modeluser.File, error)
}

// NormalizeAvatarFile 校验并清洗一次头像变更请求，返回可落库的文件 UUID。
//
// 规则与 PUT /api/v1/user/avatar 完全一致（两条路径共用本函数，避免校验强度漂移）：
//  1. UUID 非空、长度不超过 consts.MaxAvatarLen；
//  2. 文件必须存在、为 active 状态、属于当前用户、且为图片类型
//     （由 service/UserFile 校验，防止引用他人文件或非图片文件）；
//  3. 文件必须公开（permission=0），否则头像无法被所有人查看。
//
// 权限位（PermAvatar）校验由调用方通过 Update 的 Field.Perm 统一完成。
func NormalizeAvatarFile(store avatarStore, uid uint32, rawUUID string) (string, error) {
	avatarUUID := strings.TrimSpace(rawUUID)
	if avatarUUID == "" {
		return "", ErrAvatarEmpty
	}
	if len(avatarUUID) > consts.MaxAvatarLen {
		return "", ErrAvatarInvalid
	}

	file, ok := UserFile.ValidateAvatarUUID(store, uid, avatarUUID)
	if !ok || file == nil {
		return "", ErrAvatarInvalid
	}
	// 头像展示不受拉黑/屏蔽影响，必须是公开文件才能被所有人看到
	if file.Permission != consts.FilePermissionPublic {
		return "", ErrAvatarInvalid
	}

	return avatarUUID, nil
}

// AvatarField 构造头像更新字段（携带 PermAvatar 权限位，由 Update 统一校验封禁状态）。
func AvatarField(store avatarStore, uid uint32, rawUUID string) (Field, error) {
	avatarUUID, err := NormalizeAvatarFile(store, uid, rawUUID)
	if err != nil {
		return Field{}, err
	}
	return Field{Column: "avatar", Value: avatarUUID, Perm: PermAvatar}, nil
}
