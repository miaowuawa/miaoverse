-- 通知表补充「触发者 + 关联对象」字段：被点赞/被回复/被关注时，前端需要展示
-- 「谁」对「哪个对象」做了什么，并据此跳转到对应动态/评论/用户主页。
-- actor_id   触发通知的用户 id，账号安全类通知（登录/改密）无触发者，为 0；
-- target_type 关联对象类型，复用 consts.InteractTarget*（0 用户、1 动态、2 评论）；
-- target_id  关联对象 id（用户 id / 动态 id / 评论 id）。
-- 另补充 (user_id, type) 索引，支撑按分类查询与分类未读数统计。
-- MySQL 8 不支持 ADD COLUMN IF NOT EXISTS，这里用 information_schema 检查 +
-- PREPARE 动态执行，保证 db.sql 已建好列（全新库场景）时本迁移幂等跳过。

SET @has_actor_id = (SELECT COUNT(*) FROM information_schema.COLUMNS
                     WHERE TABLE_SCHEMA = DATABASE()
                       AND TABLE_NAME = 'notify'
                       AND COLUMN_NAME = 'actor_id');

SET @ddl = IF(@has_actor_id = 0,
    'ALTER TABLE `notify`
        ADD COLUMN `actor_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''trigger user id, 0 for system/account-security notify'' AFTER `type`,
        ADD COLUMN `target_type` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''0 user, 1 moment, 2 comment (consts.InteractTarget*)'' AFTER `actor_id`,
        ADD COLUMN `target_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''related target id (user id / moment id / comment id)'' AFTER `target_type`',
    'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_type_idx = (SELECT COUNT(*) FROM information_schema.STATISTICS
                     WHERE TABLE_SCHEMA = DATABASE()
                       AND TABLE_NAME = 'notify'
                       AND INDEX_NAME = 'idx_notify_user_id_type');

SET @ddl = IF(@has_type_idx = 0,
    'ALTER TABLE `notify` ADD KEY `idx_notify_user_id_type` (`user_id`, `type`)',
    'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
