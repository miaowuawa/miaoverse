package updatereq

// ProfileFull 全量更新本人基础资料。
//
// 刻意不包含 phone / region：手机号与区号属于账号身份凭据（写入 user_credentials），
// 当前不支持更改，改由专用流程处理。
//
// avatar 为可选字段（指针）：头像文件必须先上传（POST /api/v1/user/files，permission=0），
// 服务端会按与 PUT /api/v1/user/avatar 完全相同的规则校验文件归属、active 状态、
// 图片类型与公开权限，并校验 PermAvatar 权限位；为空或不传表示不修改头像。
type ProfileFull struct {
	Username string  `json:"username"`
	Nickname string  `json:"nickname"`
	Avatar   *string `json:"avatar"`
	Bio      string  `json:"bio"`
	Gender   uint8   `json:"gender"`
}

// ProfilePatch 部分更新本人基础资料，只有显式传入的字段才会被修改。
// 头像同样需要先上传文件再传 avatar（文件 UUID）。
type ProfilePatch struct {
	Username *string `json:"username"`
	Nickname *string `json:"nickname"`
	Avatar   *string `json:"avatar"`
	Bio      *string `json:"bio"`
	Gender   *uint8  `json:"gender"`
}
