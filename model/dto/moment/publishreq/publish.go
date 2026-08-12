package publishreq

// PublishMoment 发布动态请求。FileUUIDs 为图片文件 UUID 列表（可空），
// 服务端校验 UUID 合法、文件存在且为当前用户可用的 active 图片后写入关联。
type PublishMoment struct {
	Title             string   `json:"title"`
	Content           string   `json:"content"`
	Status            uint8    `json:"status"`
	Permission        uint8    `json:"permission"`
	CommentPermission uint8    `json:"comment_permission"`
	Top               uint8    `json:"top"`
	FileUUIDs         []string `json:"file_uuids"`
}
