package tag

type Tag struct {
	Id        uint64 `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Status    int    `json:"status"` // 0: 正常，1: 屏蔽

}
