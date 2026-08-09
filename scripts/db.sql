CREATE TABLE IF NOT EXISTS `user` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'user id',
    `username` VARCHAR(64) NOT NULL COMMENT 'account name',
    `nickname` VARCHAR(64) NOT NULL COMMENT 'display name',
    `region` SMALLINT UNSIGNED NOT NULL COMMENT 'phone region code',
    `avatar` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'avatar url',
    `gender` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 unknown, 1 male, 2 female, 3 non-binary',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1 active, 2 banned, 3 closed, 4 disabled',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_username` (`username`),
    UNIQUE KEY `uk_user_nickname` (`nickname`),
    KEY `idx_user_region` (`region`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='users';

CREATE TABLE IF NOT EXISTS `user_credentials` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'credential id',
    `user_id` INT UNSIGNED NOT NULL COMMENT 'user.id',
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
    `uuid`          CHAR(36)        NOT NULL COMMENT 'public file uuid',
    `user_id`       INT UNSIGNED    NOT NULL COMMENT 'uploader user id',
    `file_name`     VARCHAR(255)    NOT NULL COMMENT 'original file name',
    `object_key`    VARCHAR(500)    NOT NULL COMMENT 's3 object key',
    `file_url`      VARCHAR(500)    NOT NULL COMMENT 'file storage url',
    `file_type`     TINYINT UNSIGNED NOT NULL COMMENT '1 image, 2 video, 3 audio, 4 document, 5 other',
    `file_ext`      VARCHAR(20)     NOT NULL DEFAULT '' COMMENT 'file extension: jpg, mp4, pdf etc.',
    `mime_type`     VARCHAR(100)    NOT NULL DEFAULT '' COMMENT 'MIME type: image/jpeg, video/mp4',
    `file_size`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'file size in bytes',
    `permission`    TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '0 public, 1 friends, 2 none, 3 fans',
    `width`         INT UNSIGNED    NULL DEFAULT NULL COMMENT 'image/video width in pixels',
    `height`        INT UNSIGNED    NULL DEFAULT NULL COMMENT 'image/video height in pixels',
    `duration`      INT UNSIGNED    NULL DEFAULT NULL COMMENT 'video/audio duration in seconds',
    `thumbnail_url` VARCHAR(500)    NULL DEFAULT NULL COMMENT 'thumbnail url for video/audio',
    `hash`          BINARY(32)      NOT NULL DEFAULT 0x00 COMMENT 'sha256 hash for dedup',
    `status`        TINYINT UNSIGNED NOT NULL DEFAULT 1 
        COMMENT '1 active, 2 processing, 3 failed, 4 deleted',
    `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_files_uuid` (`uuid`),
    KEY `idx_files_user_id_status` (`user_id`, `status`),
    KEY `idx_files_file_type` (`file_type`),
    KEY `idx_files_hash` (`hash`),
    KEY `idx_files_created_at` (`created_at`),
    CONSTRAINT `fk_files_user_id` 
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='files storage';

CREATE TABLE IF NOT EXISTS `moment` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'moment id',
    `user_id`    INT UNSIGNED    NOT NULL COMMENT 'author user id',
    `title`      VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'title',
    `content`    TEXT            NOT NULL COMMENT 'content',
    `status`     TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 normal, 1 deleted, 2 draft, 3 restricted, 4 blocked',
    `permission` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 public, 1 friends, 2 private, 3 fans',
    `comment_permission` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 all, 1 friends only, 2 fans only, 3 none',
    `top`        TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 none, 1 personal top, 100 global top',
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME        NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_moment_user_id` (`user_id`),
    KEY `idx_moment_status_created` (`status`, `created_at`),
    CONSTRAINT `fk_moment_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='moments';

CREATE TABLE IF NOT EXISTS `moment_interact_count` (
    `moment_id`     BIGINT UNSIGNED NOT NULL COMMENT 'moment.id',
    `like_count`    INT UNSIGNED NOT NULL DEFAULT 0,
    `comment_count` INT UNSIGNED NOT NULL DEFAULT 0,
    `share_count`   INT UNSIGNED NOT NULL DEFAULT 0,
    `view_count`    INT UNSIGNED NOT NULL DEFAULT 0,
    `click_count`   INT UNSIGNED NOT NULL DEFAULT 0,
    `repost_count`  INT UNSIGNED NOT NULL DEFAULT 0,
    `updated_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`moment_id`),
    CONSTRAINT `fk_moment_interact_count_moment_id`
        FOREIGN KEY (`moment_id`) REFERENCES `moment` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='moment interact counters';

CREATE TABLE IF NOT EXISTS `interacts` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'interact id',
    `user_from`    INT UNSIGNED    NOT NULL COMMENT 'actor user id',
    `user_to`      INT UNSIGNED    NOT NULL COMMENT 'target user id',
    `target_id`    BIGINT UNSIGNED NOT NULL COMMENT 'target id: user/moment/comment id',
    `reference_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'reference id: comment/conversation id, 0 if none',
    `type`         TINYINT UNSIGNED NOT NULL COMMENT '0 follow, 1 like, 2 share, 3 repost, 4 favorite, 100-102 dm, 103 comment, 104 reply',
    `target_type`  TINYINT UNSIGNED NOT NULL COMMENT '0 user, 1 moment, 2 comment, 3 reply',
    `status`       TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 normal, 9 revoked, 10 forced revoked',
    `acted_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_interacts_user_from` (`user_from`),
    KEY `idx_interacts_user_to` (`user_to`),
    KEY `idx_interacts_target` (`target_id`, `type`, `status`),
    CONSTRAINT `fk_interacts_user_from`
        FOREIGN KEY (`user_from`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_interacts_user_to`
        FOREIGN KEY (`user_to`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='user interactions';

CREATE TABLE IF NOT EXISTS `comment` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'comment id',
    `user_id`     INT UNSIGNED    NOT NULL COMMENT 'author user id',
    `target_id`   BIGINT UNSIGNED NOT NULL COMMENT 'target id: moment/comment id',
    `target_type` TINYINT UNSIGNED NOT NULL COMMENT '1 moment, 2 comment (reply)',
    `content`     TEXT            NOT NULL COMMENT 'comment content',
    `status`      TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 normal, 1 deleted, 2 draft, 3 restricted, 4 blocked',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  DATETIME        NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_comment_user_id` (`user_id`),
    KEY `idx_comment_target` (`target_id`, `target_type`),
    CONSTRAINT `fk_comment_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='comments';

CREATE TABLE IF NOT EXISTS `notify` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'notify id',
    `user_id`    INT UNSIGNED    NOT NULL COMMENT 'receiver user id',
    `type`       TINYINT UNSIGNED NOT NULL COMMENT '0 account security, 1 transaction, 2 like, 3 follow, 4 mention, 5 reply/comment',
    `content`    VARCHAR(1000)   NOT NULL DEFAULT '' COMMENT 'notify content',
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `read_at`    DATETIME        NULL DEFAULT NULL,
    `received`   TINYINT(1)      NOT NULL DEFAULT 0 COMMENT 'delivered flag',
    `status`     TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 unread, 1 read, 2 deleted',
    PRIMARY KEY (`id`),
    KEY `idx_notify_user_id_status` (`user_id`, `status`),
    KEY `idx_notify_created_at` (`created_at`),
    CONSTRAINT `fk_notify_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='notifications';

CREATE TABLE IF NOT EXISTS `punishment` (
    `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'punishment record id',
    `user_id`            INT UNSIGNED    NOT NULL COMMENT 'punished user id',
    `punishment_type`    INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT 'permission bitmask: bit0 comment, bit1 post, bit2 private msg, bit3 avatar, bit4 nickname, bit5 signature, bit6 social, bit7 delete/register, bit8 upload file',
    `punishment_status`  TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1 active, 2 ended, 3 revoked',
    `punishment_time`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when punishment started',
    `punishment_end_time` DATETIME       NULL DEFAULT NULL COMMENT 'when punishment ends, NULL means permanent',
    `punishment_reason`  VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'punish reason',
    `punishment_operator` INT UNSIGNED   NOT NULL DEFAULT 0 COMMENT 'operator user id, 0 for system',
    `punishment_remark`  VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'extra remark',
    PRIMARY KEY (`id`),
    KEY `idx_punishment_user_id_status` (`user_id`, `punishment_status`),
    KEY `idx_punishment_end_time` (`punishment_end_time`),
    CONSTRAINT `fk_punishment_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='user permission punishments';

CREATE TABLE IF NOT EXISTS `article_meta` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'article id (exposed to frontend)',
    `mongo_id`    VARCHAR(24)     NOT NULL DEFAULT '' COMMENT 'mongodb article _id (hex), internal only',
    `user_id`     INT UNSIGNED    NOT NULL COMMENT 'author user id',
    `title`       VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'article title',
    `description` VARCHAR(500)    NOT NULL DEFAULT '' COMMENT 'article description/summary',
    `preview_head` VARCHAR(300)   NOT NULL DEFAULT '' COMMENT 'preview of article head',
    `cover`       VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'cover image url',
    `type`        TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 normal, 1 repost, 2 novel',
    `novel_id`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'novel root article id, 0 means not a chapter',
    `chapter_id`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'chapter number of the novel (1-based), valid for novel type',
    `status`      TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 normal, 1 deleted, 2 draft, 3 restricted, 4 blocked',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  DATETIME        NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_article_meta_mongo_id` (`mongo_id`),
    KEY `idx_article_meta_user_id_status` (`user_id`, `status`),
    KEY `idx_article_meta_novel_chapter` (`novel_id`, `chapter_id`),
    KEY `idx_article_meta_created_at` (`created_at`),
    CONSTRAINT `fk_article_meta_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='article metadata (body in mongodb)';

CREATE TABLE IF NOT EXISTS `article_interact_count` (
    `article_id`    BIGINT UNSIGNED NOT NULL COMMENT 'article_meta.id',
    `like_count`    INT UNSIGNED NOT NULL DEFAULT 0,
    `comment_count` INT UNSIGNED NOT NULL DEFAULT 0,
    `share_count`   INT UNSIGNED NOT NULL DEFAULT 0,
    `view_count`    INT UNSIGNED NOT NULL DEFAULT 0,
    `click_count`   INT UNSIGNED NOT NULL DEFAULT 0,
    `repost_count`  INT UNSIGNED NOT NULL DEFAULT 0,
    `updated_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`article_id`),
    CONSTRAINT `fk_article_interact_count_article_id`
        FOREIGN KEY (`article_id`) REFERENCES `article_meta` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='article interact counters';