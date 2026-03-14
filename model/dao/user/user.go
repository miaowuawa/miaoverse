package user

import "time"

/*
CREATE TABLE `user` (
                        `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户唯一ID',
                        `username` varchar(64) NOT NULL COMMENT '昵称（必填）',
                        `avatar` varchar(255) DEFAULT '' COMMENT '头像URL',
                        `phone` varchar(20) NOT NULL COMMENT '手机号（支持多账户绑定同一手机号）',
                        `gender` tinyint DEFAULT 0 COMMENT '0-未知 1-男 2-女 3-非二元性别',
                        `status` tinyint NOT NULL DEFAULT 1 COMMENT '1-正常 2-封禁 3-注销',
                        `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
                        `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        PRIMARY KEY (`id`),
                        KEY `idx_phone` (`phone`) COMMENT '手机号索引（方便按手机号查多账户）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户基础信息表';*/

type User struct {
	ID        uint64    `gorm:"primaryKey"`
	Username  string    `gorm:"not null"`
	Avatar    string    `gorm:"default:''"`
	Phone     string    `gorm:"not null"`
	Gender    uint8     `gorm:"default:0"`
	Status    uint8     `gorm:"not null;default:1"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (User) TableName() string {
	return "user"
}
