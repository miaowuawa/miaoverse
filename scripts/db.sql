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

CREATE TABLE IF NOT EXISTS `files` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'file id',
    `user_id`       BIGINT UNSIGNED NOT NULL COMMENT 'uploader user id',
    `file_name`     VARCHAR(255)    NOT NULL COMMENT 'original file name',
    `file_url`      VARCHAR(500)    NOT NULL COMMENT 'file storage url',
    `file_type`     VARCHAR(20)     NOT NULL COMMENT 'image, video, audio, document, other',
    `file_ext`      VARCHAR(20)     NOT NULL DEFAULT '' COMMENT 'file extension: jpg, mp4, pdf etc.',
    `mime_type`     VARCHAR(100)    NOT NULL DEFAULT '' COMMENT 'MIME type: image/jpeg, video/mp4',
    `file_size`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'file size in bytes',
    `width`         INT UNSIGNED    NULL DEFAULT NULL COMMENT 'image/video width in pixels',
    `height`        INT UNSIGNED    NULL DEFAULT NULL COMMENT 'image/video height in pixels',
    `duration`      INT UNSIGNED    NULL DEFAULT NULL COMMENT 'video/audio duration in seconds',
    `thumbnail_url` VARCHAR(500)    NULL DEFAULT NULL COMMENT 'thumbnail url for video/audio',
    `hash`          VARCHAR(64)     NOT NULL DEFAULT '' COMMENT 'file md5/sha256 hash for dedup',
    `status`        TINYINT UNSIGNED NOT NULL DEFAULT 1 
        COMMENT '1 active, 2 processing, 3 failed, 4 deleted',
    `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_files_user_id_status` (`user_id`, `status`),
    KEY `idx_files_file_type` (`file_type`),
    KEY `idx_files_hash` (`hash`),
    KEY `idx_files_created_at` (`created_at`),
    CONSTRAINT `fk_files_user_id` 
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='files storage';

