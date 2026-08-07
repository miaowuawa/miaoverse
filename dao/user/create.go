package user

import (
	"miaoverse/model/dao/user"
)

func (d *UserDAO) Create(newUser user.User, credentials []user.UserCredential) (uint32, error) {
	tx := d.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	for i := range credentials {
		credentials[i].UserID = newUser.ID
	}
	if len(credentials) > 0 {
		if err := tx.Create(&credentials).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return newUser.ID, nil
}
