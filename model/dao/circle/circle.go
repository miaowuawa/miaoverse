package circle

import "time"

type Circle struct {
	ID           uint32    `json:"id"`
	Name         string    `json:"name"`
	Owner        uint32    `json:"owner"`
	Avatar       string    `json:"avatar"`
	Description  string    `json:"description"`
	Status       uint8     `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LatestPostAt time.Time `json:"latest_post"`
	PostCount    uint32    `json:"post_count"`
}
