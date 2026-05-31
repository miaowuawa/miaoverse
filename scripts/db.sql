CREATE TABLE IF NOT EXISTS `user` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'user id',
    `username` VARCHAR(64) NOT NULL COMMENT 'account name',
    `nickname` VARCHAR(64) NOT NULL COMMENT 'display name',
    `region` INT NOT NULL COMMENT 'phone region code',
    `avatar` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'avatar url',
    `gender` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 unknown, 1 male, 2 female, 3 non-binary',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1 active, 2 banned, 3 closed',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_username` (`username`),
    UNIQUE KEY `uk_user_nickname` (`nickname`),
    KEY `idx_user_region` (`region`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='users';

CREATE TABLE IF NOT EXISTS `user_credentials` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'credential id',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'user.id',
    `credential_type` TINYINT UNSIGNED NOT NULL COMMENT '1 password, 2 phone, 3 third-party/webauthn',
    `credential_key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'credential key',
    `credential_value` VARCHAR(512) NOT NULL COMMENT 'credential value',
    `credential_ext` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'extra json data',
    `last_login_at` DATETIME NULL DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_credential_type_key` (`user_id`, `credential_type`, `credential_key`),
    KEY `idx_credential_type_key_value` (`credential_type`, `credential_key`, `credential_value`),
    CONSTRAINT `fk_user_credentials_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='user credentials';
