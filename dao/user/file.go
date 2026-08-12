package user

import (
	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
)

func (d *UserDAO) CreateFile(file modeluser.File) (*modeluser.File, error) {
	if err := d.DB.Create(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (d *UserDAO) QueryActiveFileByUUID(userID uint32, fileUUID string) (*modeluser.File, error) {
	var file modeluser.File
	err := d.DB.Where("user_id = ? AND uuid = ? AND status = ?", userID, fileUUID, consts.FileStatusActive).First(&file).Error
	return &file, err
}

// QueryActiveFileByUUIDAnyUser 按 UUID 查询任意用户的 active 文件，用于帖子等场景查看/下载他人文件
func (d *UserDAO) QueryActiveFileByUUIDAnyUser(fileUUID string) (*modeluser.File, error) {
	var file modeluser.File
	err := d.DB.Where("uuid = ? AND status = ?", fileUUID, consts.FileStatusActive).First(&file).Error
	return &file, err
}

// QueryActiveFilesByUUIDsBatch 按 UUID 列表批量查询 active 文件（发布动态图片校验场景）。
// 返回 uuid → 文件记录；不存在的 UUID 不会出现在结果中。
func (d *UserDAO) QueryActiveFilesByUUIDsBatch(fileUUIDs []string) (map[string]modeluser.File, error) {
	result := map[string]modeluser.File{}
	if len(fileUUIDs) == 0 {
		return result, nil
	}
	var list []modeluser.File
	if err := d.DB.Where("uuid IN ? AND status = ?", fileUUIDs, consts.FileStatusActive).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, file := range list {
		result[file.UUID] = file
	}
	return result, nil
}

func (d *UserDAO) QueryActiveFileByHash(hash [32]byte) (*modeluser.File, error) {
	var file modeluser.File
	err := d.DB.Where("hash = ? AND status = ?", hash[:], consts.FileStatusActive).Order("id ASC").First(&file).Error
	return &file, err
}
