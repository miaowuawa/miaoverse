package updatereq

// UpdateMoment 编辑动态请求。所有字段均为指针，nil 表示不修改该字段（部分更新）。
type UpdateMoment struct {
	Title             *string `json:"title"`
	Content           *string `json:"content"`
	Status            *uint8  `json:"status"`
	Permission        *uint8  `json:"permission"`
	CommentPermission *uint8  `json:"comment_permission"`
	Top               *uint8  `json:"top"`
}
