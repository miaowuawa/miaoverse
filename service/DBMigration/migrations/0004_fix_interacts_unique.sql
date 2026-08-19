-- 修复并发点赞/关注等单实例互动的竞态问题。
-- 背景：interacts 表此前没有唯一约束，"先查再插"的点赞/关注在并发下会对同一
-- (user_from, target_id, type, target_type) 插入多行，导致互动状态重复、计数漂移。

-- 1. 去重：按单实例互动类型（0 关注 / 1 点赞 / 2 分享 / 3 转发 / 4 收藏）分组，
--    每个分组只保留一行，优先保留正常（status=0）行，其次保留 id 最小行。
--    必须先去重，否则第 2 步的唯一索引会因历史重复数据创建失败。
--    外层再包一层派生表是 MySQL 同表删除的标准写法（error 1093 规避），
--    ROW_NUMBER() 会强制派生表物化，不受派生表合并优化影响。
DELETE FROM `interacts`
WHERE `id` IN (
    SELECT `id` FROM (
        SELECT `id`,
               ROW_NUMBER() OVER (
                   PARTITION BY `user_from`, `target_id`, `type`, `target_type`
                   ORDER BY (`status` = 0) DESC, `id` ASC
               ) AS `rn`
        FROM `interacts`
        WHERE `type` IN (0, 1, 2, 3, 4)
    ) AS `ranked`
    WHERE `rn` > 1
);

-- 2. 添加生成列 + 唯一索引，从数据库层面彻底消除单实例互动的重复行竞态。
--    评论（103）/回复（104）/私信（100-102）允许同一用户对同一目标多次操作，
--    single_key 为 NULL（NULL 不参与唯一约束），因此不能用整表四列唯一索引。
--    MySQL 8 不支持 ADD COLUMN IF NOT EXISTS，这里用 information_schema 检查 +
--    PREPARE 动态执行，保证 db.sql 已建好列（全新库场景）时本迁移幂等跳过。
SET @has_single_key = (SELECT COUNT(*) FROM information_schema.COLUMNS
                       WHERE TABLE_SCHEMA = DATABASE()
                         AND TABLE_NAME = 'interacts'
                         AND COLUMN_NAME = 'single_key');

SET @ddl = IF(@has_single_key = 0, 'ALTER TABLE `interacts` ADD COLUMN `single_key` VARCHAR(64) GENERATED ALWAYS AS (IF(`type` IN (0, 1, 2, 3, 4), CONCAT(`user_from`, '':'', `target_id`, '':'', `type`, '':'', `target_type`), NULL)) STORED COMMENT ''unique key for single-instance interactions (follow/like/share/repost/favorite), NULL for multi-instance types'', ADD UNIQUE KEY `uk_interacts_single_key` (`single_key`)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. 按实际正常点赞行重算点赞计数，纠正历史竞态导致的计数漂移。
--    动态与文章点赞统一复用 target_type=1（与 HasLikedMoment/HasLikedArticlesBatch 口径一致）。
UPDATE `moment_interact_count` mic
JOIN (
    SELECT `target_id`, COUNT(*) AS `c`
    FROM `interacts`
    WHERE `type` = 1 AND `target_type` = 1 AND `status` = 0
    GROUP BY `target_id`
) t ON t.`target_id` = mic.`moment_id`
SET mic.`like_count` = t.`c`;

UPDATE `comment_interact_count` cic
JOIN (
    SELECT `target_id`, COUNT(*) AS `c`
    FROM `interacts`
    WHERE `type` = 1 AND `target_type` = 2 AND `status` = 0
    GROUP BY `target_id`
) t ON t.`target_id` = cic.`comment_id`
SET cic.`like_count` = t.`c`;