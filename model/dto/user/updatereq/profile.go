package updatereq

type ProfileFull struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Region   uint16 `json:"region"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Gender   uint8  `json:"gender"`
}

type ProfilePatch struct {
	Username *string `json:"username"`
	Nickname *string `json:"nickname"`
	Region   *uint16 `json:"region"`
	Avatar   *string `json:"avatar"`
	Bio      *string `json:"bio"`
	Gender   *uint8  `json:"gender"`
}
