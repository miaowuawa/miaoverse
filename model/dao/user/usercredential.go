package user

import "time"

/*
CREATE TABLE `user_credentials` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '凭证ID',
    `user_id` INT UNSIGNED NOT NULL COMMENT '关联user.id',
    `credential_type` TINYINT UNSIGNED NOT NULL COMMENT '1-密码 2-手机验证 3-第三方登录',
    `credential_key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '通行密钥ID/手机号/第三方平台（如wechat）',
    `credential_value` VARCHAR(512) NOT NULL COMMENT '密码哈希/通行密钥公钥/第三方openid',
    `credential_ext` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '扩展字段（JSON格式，存非核心参数）',
    `last_login_at` DATETIME NULL DEFAULT NULL COMMENT '最后登录时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id_type_key` (`user_id`,`credential_type`,`credential_key`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_type_key` (`credential_type`,`credential_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户凭证/密码表';
*/

type UserCredential struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	UserID          uint32     `gorm:"not null" json:"user_id"`
	CredentialType  uint8      `gorm:"not null" json:"credential_type"`
	CredentialKey   string     `gorm:"not null" json:"credential_key"`
	CredentialValue string     `gorm:"not null" json:"credential_value"`
	CredentialExt   string     `gorm:"default:''" json:"credential_ext"`
	LastLoginAt     *time.Time `gorm:"default:null" json:"last_login_at"`
	CreatedAt       time.Time  `gorm:"not null;default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"not null;default:current_timestamp on update current_timestamp" json:"updated_at"`
}

func (UserCredential) TableName() string {
	return "user_credentials"
}
