package user

import "time"

/*
CREATE TABLE `user` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户唯一ID',
    `username` VARCHAR(64) NOT NULL COMMENT '账号名（唯一）',
    `nickname` VARCHAR(64) NOT NULL COMMENT '昵称（唯一）',
    `region` SMALLINT UNSIGNED NOT NULL COMMENT '手机区号，如 86',
    `avatar` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像URL',
    `gender` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0-未知 1-男 2-女 3-非二元性别',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1-正常 2-封禁 3-注销 4-停用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_username` (`username`),
    UNIQUE KEY `uk_user_nickname` (`nickname`),
    KEY `idx_user_region` (`region`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户基础信息表';
*/

type User struct {
	ID        uint32    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"not null;uniqueIndex" json:"username"`
	Nickname  string    `gorm:"not null;uniqueIndex" json:"nickname"`
	Region    uint16    `gorm:"not null" json:"region"`
	Avatar    string    `gorm:"default:''" json:"avatar"`
	Gender    uint8     `gorm:"default:0" json:"gender"`
	Status    uint8     `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}
