CREATE TABLE `user` (
                        `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户唯一ID',
                        `username` varchar(64) NOT NULL COMMENT '昵称（必填）',
                        `avatar` varchar(255) DEFAULT '' COMMENT '头像URL',
                        `phone` varchar(20) NOT NULL COMMENT '手机号（支持多账户绑定同一手机号）',
                        `gender` tinyint DEFAULT 0 COMMENT '0-未知 1-男 2-女 3-非二元性别',
                        `status` tinyint NOT NULL DEFAULT 1 COMMENT '1-正常 2-封禁 3-注销',
                        `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
                        `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        PRIMARY KEY (`id`),
                        KEY `idx_phone` (`phone`) COMMENT '手机号索引（方便按手机号查多账户）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户基础信息表';

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
                                    CONSTRAINT `fk_uc_user_id` FOREIGN KEY (`user_id`) 
                                        REFERENCES `user` (`id`) 
                                        ON DELETE CASCADE  -- 级联删除：删用户时自动删其凭证（重点！）
                                        ON UPDATE CASCADE  -- 级联更新：若user.id变更（极少用），自动同步
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户凭证/密码表';