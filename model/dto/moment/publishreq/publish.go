package publishreq

type PublishMoment struct {
	Title             string `json:"title"`
	Content           string `json:"content"`
	Status            uint8  `json:"status"`
	Permission        uint8  `json:"permission"`
	CommentPermission uint8  `json:"comment_permission"`
	Top               uint8  `json:"top"`
}
