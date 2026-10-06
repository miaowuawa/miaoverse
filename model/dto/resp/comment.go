package resp

import (
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
// Content 中可能包含贴纸内嵌标记 [sticker:<uuid>]（作为文字的一部分穿插展示），
// Sticker 为该贴纸的展示信息（含 Hidden 封禁/失效提示），无贴纸时为 nil。
// 原始存储 URL 不下发，贴纸图片经临时链接接口用 FileUUID 换取。
type CommentInfo struct {
	ID        uint64          `json:"id"`
	UserID    uint32          `json:"user_id"`
	MomentID  uint64          `json:"moment_id"`
	Content   string          `json:"content"`
	Status    uint8           `json:"status"`
	CreatedAt string          `json:"created_at"`
	Author    modeluser.User  `json:"author"`
	Likes     uint32          `json:"likes"`
	IsLiked   bool            `json:"is_liked"`
	Sticker   *CommentSticker `json:"sticker,omitempty"`
}

// ReplyInfo 楼中楼回复信息。MomentID 为所属动态 ID，ReplyToID/ReplyToUserID 为被回复的评论及其作者。
// Sticker 语义与 CommentInfo 一致。
type ReplyInfo struct {
	ID            uint64          `json:"id"`
	UserID        uint32          `json:"user_id"`
	MomentID      uint64          `json:"moment_id"`
	ReplyToID     uint64          `json:"reply_to_id"`
	ReplyToUserID uint32          `json:"reply_to_user_id"`
	Content       string          `json:"content"`
	Status        uint8           `json:"status"`
	CreatedAt     string          `json:"created_at"`
	Sticker       *CommentSticker `json:"sticker,omitempty"`
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
