package commentreq

type CreateComment struct {
	MomentID uint64 `json:"moment_id"`
	Content  string `json:"content"`
}
