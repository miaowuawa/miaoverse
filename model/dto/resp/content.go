package resp

type ContentItem struct {
	ID           uint64 `json:"id"`
	Type         string `json:"type"`
	Comment      uint32 `json:"comment"`
	Like         uint32 `json:"like"`
	ChapterCount uint64 `json:"chapter_count"` // 小说根文章的已发布章节数，其余类型为 0
}

type CodeWithMsgContentList struct {
	Code     int           `json:"code"`
	Msg      string        `json:"msg"`
	Count    int64         `json:"count"`
	Contents []ContentItem `json:"contents"`
}
