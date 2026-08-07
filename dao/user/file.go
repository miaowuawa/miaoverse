package user

import modeluser "miaoverse/model/dao/user"

func (d *UserDAO) CreateFile(file modeluser.File) (*modeluser.File, error) {
	if err := d.DB.Create(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (d *UserDAO) QueryActiveFileByUUID(userID uint32, fileUUID string) (*modeluser.File, error) {
	var file modeluser.File
	err := d.DB.Where("user_id = ? AND uuid = ? AND status = ?", userID, fileUUID, modeluser.FileStatusActive).First(&file).Error
	return &file, err
}

// QueryActiveFileByUUIDAnyUser 按 UUID 查询任意用户的 active 文件，用于帖子等场景查看/下载他人文件
func (d *UserDAO) QueryActiveFileByUUIDAnyUser(fileUUID string) (*modeluser.File, error) {
	var file modeluser.File
	err := d.DB.Where("uuid = ? AND status = ?", fileUUID, modeluser.FileStatusActive).First(&file).Error
	return &file, err
}

func (d *UserDAO) QueryActiveFileByHash(hash [32]byte) (*modeluser.File, error) {
	var file modeluser.File
	err := d.DB.Where("hash = ? AND status = ?", hash[:], modeluser.FileStatusActive).Order("id ASC").First(&file).Error
	return &file, err
}
