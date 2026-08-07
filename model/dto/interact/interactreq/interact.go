package interactreq

type FollowUser struct {
	Target uint32 `json:"target"`
}

type LikeMoment struct {
	MomentID uint64 `json:"moment_id"`
}
