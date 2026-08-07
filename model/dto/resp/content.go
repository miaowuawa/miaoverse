package resp

type ContentItem struct {
	ID      uint64 `json:"id"`
	Type    string `json:"type"`
	Comment uint32 `json:"comment"`
	Like    uint32 `json:"like"`
}

type CodeWithMsgContentList struct {
	Code     int           `json:"code"`
	Msg      string        `json:"msg"`
	Count    int64         `json:"count"`
	Contents []ContentItem `json:"contents"`
}
