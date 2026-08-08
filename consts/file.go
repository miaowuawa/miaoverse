package consts

// 文件域常量：状态、类型、分享权限、上传限制。

const (
	FileStatusActive     uint8 = 1
	FileStatusProcessing uint8 = 2
	FileStatusFailed     uint8 = 3
	FileStatusDeleted    uint8 = 4
)

// FileType 文件大类分类（固定值，100% 写死不变，可用 uint8 常量）
const (
	FileTypeImage    uint8 = 1
	FileTypeVideo    uint8 = 2
	FileTypeAudio    uint8 = 3
	FileTypeDocument uint8 = 4
	FileTypeOther    uint8 = 5
)

// FilePermission 文件分享权限（与动态可见权限一致）
const (
	FilePermissionPublic  uint8 = 0 // 给全部人公开
	FilePermissionFriends uint8 = 1 // 给好友公开
	FilePermissionFans    uint8 = 3 // 给粉丝公开
	FilePermissionNone    uint8 = 2 // 不给任何人公开
)

const (
	FormFileField         = "file"
	MultipartBodyOverhead = 1024 * 1024 // multipart 请求 body 额外开销（字节）
)
