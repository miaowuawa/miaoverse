package passwordreq

// ChangePassword 通过手机验证码设置/修改当前登录账号的密码。
// 注意：请求体刻意不接收手机号与区号——两者一律取自当前登录会话，
// 避免越权修改他人账号密码，也符合「手机号不支持更改」的约束。
type ChangePassword struct {
	// 修改密码场景下发的验证码 UUID（POST /api/v1/auth/sms/send，action=CHANGE_PASSWORD）
	UUID string `json:"uuid" validate:"required,isUUIDv4"`
	// 短信验证码
	Code int `json:"code" validate:"required,numeric"`
	// 新密码明文，仅用于本次请求，服务端 bcrypt 哈希后落库
	Password string `json:"password" validate:"required"`
	// a 参数：当前毫秒时间戳平方后反转的十进制字符串，见 API 文档
	Timestamp string `json:"a" validate:"required,isDigit"`
}
