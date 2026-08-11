package resp

import (
	"time"

	modeluser "miaoverse/model/dao/user"
)

type MomentInfo struct {
	ID                uint64    `json:"id"`
	UserID            uint32    `json:"user_id"`
	Title             string    `json:"title"`
	Content           string    `json:"content"`
	Status            uint8     `json:"status"`
	Permission        uint8     `json:"permission"`
	CommentPermission uint8     `json:"comment_permission"`
	Top               uint8     `json:"top"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CodeWithMsgMoment struct {
	Code   int        `json:"code"`
	Msg    string     `json:"msg"`
	Moment MomentInfo `json:"moment"`
}

// MomentStats 动态互动计数（面向详情页展示）
type MomentStats struct {
	Likes    uint32 `json:"likes"`
	Comments uint32 `json:"comments"`
	Shares   uint32 `json:"shares"`
}

// MomentDetail 动态详情响应：动态本体 + 作者信息 + 互动计数 + 当前用户互动状态
type MomentDetail struct {
	ID                uint64         `json:"id"`
	UserID            uint32         `json:"user_id"`
	Title             string         `json:"title"`
	Content           string         `json:"content"`
	Status            uint8          `json:"status"`
	Permission        uint8          `json:"permission"`
	CommentPermission uint8          `json:"comment_permission"`
	Top               uint8          `json:"top"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	Author            modeluser.User `json:"author"`
	Stats             MomentStats    `json:"stats"`
	IsLiked           bool           `json:"is_liked"`
	IsFollowing       bool           `json:"is_following"`
}

type CodeWithMsgMomentDetail struct {
	Code   int          `json:"code"`
	Msg    string       `json:"msg"`
	Moment MomentDetail `json:"moment"`
}
