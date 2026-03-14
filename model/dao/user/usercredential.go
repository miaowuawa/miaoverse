package user

import "time"

/*
CREATE TABLE `user_credentials` (
                                    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '凭证ID',
                                    `user_id` bigint unsigned NOT NULL COMMENT '关联user.id',
                                    `credential_type` tinyint NOT NULL COMMENT '1-密码 3-第三方登录',
                                    `credential_key` varchar(128) DEFAULT '' COMMENT '通行密钥ID/手机号/第三方平台（如wechat）',
                                    `credential_value` varchar(512) NOT NULL COMMENT '密码哈希/通行密钥公钥/第三方openid',
                                    `credential_ext` varchar(512) DEFAULT '' COMMENT '扩展字段（JSON格式，存非核心参数）',
                                    `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
                                    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                    UNIQUE KEY `uk_user_id_type_key` (`user_id`,`credential_type`,`credential_key`) COMMENT '唯一约束：一个用户同类型凭证只能绑定一个key（如一个手机号）',
                                    KEY `idx_user_id` (`user_id`),
                                    KEY `idx_type_key` (`credential_type`,`credential_key`) COMMENT '按类型+key查（如手机号查所有用户）'
                                    PRIMARY KEY (`id`),
                                    UNIQUE KEY `uk_user_id_type` (`user_id`,`credential_type`) COMMENT '一个用户一种凭证类型只能有一条',
                                    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户凭证/密码表';*/

type UserCredential struct {
	ID              uint64    `gorm:"primaryKey"`                                                     // 凭证ID
	UserID          uint64    `gorm:"not null"`                                                       // 关联user.id
	CredentialType  uint8     `gorm:"not null"`                                                       // 凭证类型 1-密码 2-手机验证 3-第三方登录
	CredentialKey   string    `gorm:"not null"`                                                       // 通行密钥ID/手机号/第三方平台（如wechat）
	CredentialValue string    `gorm:"not null"`                                                       // 凭证值 密码bcrypt/第三方openid等
	CredentialExt   string    `gorm:"default:''"`                                                     // 扩展字段（JSON格式，存非核心参数）
	LastLoginAt     time.Time `gorm:"default:null"`                                                   //	最后登录时间
	CreatedAt       time.Time `gorm:"not null;default:current_timestamp"`                             // 创建时间
	UpdatedAt       time.Time `gorm:"not null;default:current_timestamp on update current_timestamp"` // 更新时间
}

func (UserCredential) TableName() string {
	return "user_credentials"
}
