package resp

import (
	"time"

	modeluser "miaoverse/model/dao/user"
)

type CodeWithMsgComment struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Comment CommentInfo `json:"comment"`
}

// CodeWithMsgCommentList 评论列表响应体（动态一级评论分页列表）。
type CodeWithMsgCommentList struct {
	Code     int           `json:"code"`
	Msg      string        `json:"msg"`
	Count    int64         `json:"count"`
	Comments []CommentInfo `json:"comments"`
}

// CommentInfo 评论信息。
// Content 中可能包含贴纸内嵌标记 [sticker:<uuid>]（作为文字的一部分穿插展示，一条评论最多 25 个标记），
// Stickers 为各标记位置上的贴纸展示信息（按标记出现顺序，含 Hidden 封禁/失效提示），无贴纸时为空。
// 原始存储 URL 不下发，贴纸图片经临时链接接口用 FileUUID 换取。
type CommentInfo struct {
	ID        uint64         `json:"id"`
	UserID    uint32         `json:"user_id"`
	MomentID  uint64         `json:"moment_id"`
	Content   string         `json:"content"`
	Status    uint8          `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	Author    modeluser.User `json:"author"`
	Likes     uint32         `json:"likes"`
	IsLiked   bool           `json:"is_liked"`
	// ReplyCount 为该评论楼中楼下的回复总数（含全部子孙回复）。
	ReplyCount int64 `json:"reply_count"`
	// ReplyDepth 为该评论楼中楼相对首条评论的最大嵌套层数（首条评论为 0 层，直接回复为 1 层）。
	// 前端据此决定是否在首条评论处提供「查看完整对话」入口（超过预览层数时）。
	ReplyDepth int              `json:"reply_depth"`
	Stickers   []CommentSticker `json:"stickers,omitempty"`
}

// ReplyInfo 楼中楼回复信息。MomentID 为所属动态 ID，ReplyToID/ReplyToUserID 为被回复的评论及其作者。
// Stickers 语义与 CommentInfo 一致。
type ReplyInfo struct {
	ID            uint64           `json:"id"`
	UserID        uint32           `json:"user_id"`
	MomentID      uint64           `json:"moment_id"`
	ReplyToID     uint64           `json:"reply_to_id"`
	ReplyToUserID uint32           `json:"reply_to_user_id"`
	Content       string           `json:"content"`
	Status        uint8            `json:"status"`
	CreatedAt     time.Time        `json:"created_at"`
	Author        modeluser.User   `json:"author"`
	Likes         uint32           `json:"likes"`
	IsLiked       bool             `json:"is_liked"`
	Stickers      []CommentSticker `json:"stickers,omitempty"`
}

// ConversationInfo 楼中楼完整对话：Root 为传入的楼中楼首条评论，Replies 为其全部子孙回复（扁平列表，按时间正序）。
type ConversationInfo struct {
	Root    CommentInfo `json:"root"`
	Count   int64       `json:"count"`
	Replies []ReplyInfo `json:"replies"`
}

type CodeWithMsgReply struct {
	Code  int       `json:"code"`
	Msg   string    `json:"msg"`
	Reply ReplyInfo `json:"reply"`
}

type CodeWithMsgConversation struct {
	Code         int              `json:"code"`
	Msg          string           `json:"msg"`
	Conversation ConversationInfo `json:"conversation"`
}
