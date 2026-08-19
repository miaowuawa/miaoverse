package resp

type AvatarInfo struct {
	AvatarUUID string `json:"avatar_uuid"`
}

type CodeWithMsgAvatar struct {
	Code   int        `json:"code"`
	Msg    string     `json:"msg"`
	Avatar AvatarInfo `json:"avatar"`
}
