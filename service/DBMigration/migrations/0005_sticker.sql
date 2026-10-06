-- 贴纸（stickers）功能：
-- 1. 新增 sticker（贴纸）、sticker_pack（贴纸包）、user_sticker（用户贴纸收藏夹）、
--    sticker_pack_favorite（贴纸包收藏）四张表；
-- 2. comment 表新增 sticker_uuid 列（评论内嵌贴纸，一条评论最多一个）。
--
-- 说明：
-- - 全新库场景下 db.sql 已建好这些表/列，本迁移用 IF NOT EXISTS 与 information_schema 检查保证幂等跳过；
-- - 贴纸包封禁标记 sticker_pack.banned 为运营处置字段（与内容屏蔽状态一致），置 1 后包内贴纸
--   在所有使用处（评论等）无法显示，前端在评论下灰字提示「部分贴纸未显示」。

CREATE TABLE IF NOT EXISTS `sticker` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'sticker id',
    `uuid`        CHAR(36)        NOT NULL COMMENT 'public sticker uuid',
    `user_id`     INT UNSIGNED    NOT NULL COMMENT 'uploader user id',
    `pack_id`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'sticker_pack.id, 0 if not in any pack',
    `file_uuid`   CHAR(36)        NOT NULL COMMENT 'image file uuid in files table',
    `name`        VARCHAR(64)     NOT NULL DEFAULT '' COMMENT 'sticker display name',
    `sort`        INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT 'display order within pack, ascending',
    `status`      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1 active, 4 deleted',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sticker_uuid` (`uuid`),
    KEY `idx_sticker_user_status` (`user_id`, `status`),
    KEY `idx_sticker_pack_status` (`pack_id`, `status`),
    CONSTRAINT `fk_sticker_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='stickers';

CREATE TABLE IF NOT EXISTS `sticker_pack` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'sticker pack id',
    `uuid`        CHAR(36)        NOT NULL COMMENT 'public pack uuid',
    `user_id`     INT UNSIGNED    NOT NULL COMMENT 'creator user id',
    `name`        VARCHAR(64)     NOT NULL DEFAULT '' COMMENT 'pack name',
    `description` VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'pack description',
    `banned`      TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'ban flag: 0 normal, 1 banned (stickers hidden everywhere)',
    `status`      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1 active, 4 deleted',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sticker_pack_uuid` (`uuid`),
    KEY `idx_sticker_pack_user_id` (`user_id`),
    KEY `idx_sticker_pack_banned_status` (`banned`, `status`),
    CONSTRAINT `fk_sticker_pack_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='sticker packs';

CREATE TABLE IF NOT EXISTS `user_sticker` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'user sticker id',
    `user_id`     INT UNSIGNED    NOT NULL COMMENT 'user.id',
    `sticker_id`  BIGINT UNSIGNED NOT NULL COMMENT 'sticker.id',
    `source`      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1 own upload (auto), 2 favorited from others',
    `top`         TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0 none, 1 pinned in personal collection',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_sticker_user_sticker` (`user_id`, `sticker_id`),
    KEY `idx_user_sticker_user_top` (`user_id`, `top`),
    CONSTRAINT `fk_user_sticker_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_user_sticker_sticker_id`
        FOREIGN KEY (`sticker_id`) REFERENCES `sticker` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='user sticker collection (own uploads + favorites)';

CREATE TABLE IF NOT EXISTS `sticker_pack_favorite` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'favorite id',
    `user_id`     INT UNSIGNED    NOT NULL COMMENT 'user.id',
    `pack_id`     BIGINT UNSIGNED NOT NULL COMMENT 'sticker_pack.id',
    `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sticker_pack_favorite` (`user_id`, `pack_id`),
    CONSTRAINT `fk_sticker_pack_favorite_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_sticker_pack_favorite_pack_id`
        FOREIGN KEY (`pack_id`) REFERENCES `sticker_pack` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='sticker pack favorites';

-- comment 表新增 sticker_uuid 列：MySQL 8 不支持 ADD COLUMN IF NOT EXISTS，
-- 用 information_schema 检查 + PREPARE 动态执行，保证 db.sql 已建好列（全新库场景）时幂等跳过。
SET @has_sticker_uuid = (SELECT COUNT(*) FROM information_schema.COLUMNS
                         WHERE TABLE_SCHEMA = DATABASE()
                           AND TABLE_NAME = 'comment'
                           AND COLUMN_NAME = 'sticker_uuid');

SET @ddl = IF(@has_sticker_uuid = 0, 'ALTER TABLE `comment` ADD COLUMN `sticker_uuid` CHAR(36) NOT NULL DEFAULT '''' COMMENT ''inline sticker uuid, empty if none (max 1 per comment)'' AFTER `content`', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
