-- 0006_comment_sticker_multi.sql
-- 评论贴纸规则升级：一条评论最多 25 张贴纸（此前最多 1 张）。
-- 完整贴纸列表以 comment.content 中的 [sticker:<uuid>] 内嵌标记为准（按标记出现顺序展示）；
-- sticker_uuid 列保留为「首张贴纸」兼容字段，仅同步更新列注释，不改动表结构与数据。
-- MODIFY COLUMN 为同类型元数据变更，可重复执行（幂等）。

ALTER TABLE `comment`
    MODIFY COLUMN `sticker_uuid` CHAR(36) NOT NULL DEFAULT '' COMMENT 'first inline sticker uuid (compat), empty if none; up to 25 inline sticker tokens in content';
