package UserCheck

import (
	user "miaoverse/model/dao/user"
	"miaoverse/model/server"
	phone2 "miaoverse/service/security/credentials/phone"
	"miaoverse/util"
)

func CheckAndCreateIfNotExists(phone string, region int, servants *server.Servants) ([]user.User, bool, error) {
	users, err := servants.UserServant.QueryByPhone(phone, region)
	if err != nil {
		return nil, false, err
	}
	if len(users) > 0 {
		return users, true, nil
	}

	newUser, err := CreateAccountForPhone(phone, region, servants)
	if err != nil {
		return nil, false, err
	}
	return []user.User{newUser}, false, nil
}

func CreateAccountForPhone(phone string, region int, servants *server.Servants) (user.User, error) {
	username, err := util.BuildString.GenerateUsername()
	if err != nil {
		return user.User{}, err
	}

	newUser := user.User{
		Username: username,
		Nickname: username,
		Region:   region,
	}
	credentials, err := phone2.PhoneToCredentialStructure(phone, region)
	if err != nil {
		return user.User{}, err
	}

	id, err := servants.UserServant.Create(newUser, credentials)
	if err != nil {
		return user.User{}, err
	}
	newUser.ID = id

	return newUser, nil
}

func UserBelongsToPhone(userID uint64, phone string, region int, servants *server.Servants) (bool, error) {
	users, err := servants.UserServant.QueryByPhone(phone, region)
	if err != nil {
		return false, err
	}
	for _, u := range users {
		if u.ID == userID {
			return true, nil
		}
	}
	return false, nil
}
