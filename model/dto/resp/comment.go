package resp

type CodeWithMsgComment struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Comment CommentInfo `json:"comment"`
}

type CommentInfo struct {
	ID        uint64 `json:"id"`
	UserID    uint32 `json:"user_id"`
	MomentID  uint64 `json:"moment_id"`
	Content   string `json:"content"`
	Status    uint8  `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ReplyInfo 楼中楼回复信息。MomentID 为所属动态 ID，ReplyToID/ReplyToUserID 为被回复的评论及其作者。
type ReplyInfo struct {
	ID            uint64 `json:"id"`
	UserID        uint32 `json:"user_id"`
	MomentID      uint64 `json:"moment_id"`
	ReplyToID     uint64 `json:"reply_to_id"`
	ReplyToUserID uint32 `json:"reply_to_user_id"`
	Content       string `json:"content"`
	Status        uint8  `json:"status"`
	CreatedAt     string `json:"created_at"`
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
