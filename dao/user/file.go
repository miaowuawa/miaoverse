package user

import modeluser "miaoverse/model/dao/user"

func (d *UserDAO) CreateFile(file modeluser.File) (*modeluser.File, error) {
	if err := d.DB.Create(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (d *UserDAO) QueryActiveFileByUUID(userID uint64, fileUUID string) (*modeluser.File, error) {
	var file modeluser.File
	err := d.DB.Where("user_id = ? AND uuid = ? AND status = ?", userID, fileUUID, modeluser.FileStatusActive).First(&file).Error
	return &file, err
}
