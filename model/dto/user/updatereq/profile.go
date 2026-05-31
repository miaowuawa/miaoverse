package updatereq

type ProfileFull struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Region   int    `json:"region"`
	Avatar   string `json:"avatar"`
	Gender   uint8  `json:"gender"`
}

type ProfilePatch struct {
	Username *string `json:"username"`
	Nickname *string `json:"nickname"`
	Region   *int    `json:"region"`
	Avatar   *string `json:"avatar"`
	Gender   *uint8  `json:"gender"`
}
