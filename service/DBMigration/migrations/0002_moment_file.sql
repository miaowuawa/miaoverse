CREATE TABLE IF NOT EXISTS `moment_file` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'moment file id',
    `moment_id`  BIGINT UNSIGNED NOT NULL COMMENT 'moment.id',
    `file_uuid`  CHAR(36)        NOT NULL COMMENT 'public file uuid in files table',
    `sort`       INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT 'display order, ascending',
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_moment_file_moment_sort` (`moment_id`, `sort`),
    CONSTRAINT `fk_moment_file_moment_id`
        FOREIGN KEY (`moment_id`) REFERENCES `moment` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='moment image associations';
