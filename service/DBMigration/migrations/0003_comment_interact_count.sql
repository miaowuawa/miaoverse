CREATE TABLE IF NOT EXISTS `comment_interact_count` (
    `comment_id` BIGINT UNSIGNED NOT NULL COMMENT 'comment.id',
    `like_count` INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT 'like count',
    `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`comment_id`),
    CONSTRAINT `fk_comment_interact_count_comment_id`
        FOREIGN KEY (`comment_id`) REFERENCES `comment` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='comment interact counters';
