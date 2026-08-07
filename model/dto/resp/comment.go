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
