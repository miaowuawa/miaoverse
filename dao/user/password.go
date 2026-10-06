package user

import (
	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"

	"gorm.io/gorm/clause"
)

// UpsertPassword 写入或更新用户密码凭证。
// 唯一键为 (user_id, credential_type, credential_key)，冲突时只覆盖 credential_value，
// 单条 SQL 完成写入，避免并发修改密码时产生重复凭证或读到中间状态。
// 保留 credential_key 由调用方（service/UserPassword）经凭证构造函数生成，保证与注册流程一致。
func (d *UserDAO) UpsertPassword(credential *modeluser.UserCredential) error {
	if credential == nil {
		return nil
	}
	credential.CredentialType = consts.Password

	return d.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "credential_type"}, {Name: "credential_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"credential_value"}),
	}).Create(credential).Error
}
