package resp

import "time"

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
