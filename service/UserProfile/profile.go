// Package UserProfile 负责本人基础资料（账号名、昵称、个性签名、性别）的
// 清洗、校验与落库规则。
//
// 安全边界：
//   - 资料字段只作用于当前登录用户，userID 一律由登录会话提供，本包不接受目标用户参数；
//   - 输入统一清洗（去首尾空白、统一换行、拒绝控制字符与 Unicode 格式字符），
//     避免不可见字符被用于仿冒他人或污染其它端展示；
//   - 头像与手机号/区号不属于本包职责：头像必须走 PUT /api/v1/user/avatar 的
//     文件归属与公开性校验，手机号与区号当前不支持更改。
package UserProfile

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
)

var (
	// ErrPunished 用户存在生效中的权限封禁，对应字段不可修改（响应 403 + 业务码 40302）。
	ErrPunished = errors.New("profile update rejected by active punishment")
	// ErrNoFields 没有任何可更新字段（PATCH 空请求体）。
	ErrNoFields = errors.New("no updatable profile field")
)

// usernamePattern 账号名只允许 ASCII 字母/数字/下划线/中划线/点，且必须以字母或数字开头。
// 收紧字符集的原因：账号名是用户唯一句柄，允许 Unicode 会引入同形字（confusable）仿冒，
// 以及零宽字符、双向控制符等不可见字符。
var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// 修改资料各字段所需的权限位（存在生效中封禁时不可修改）。
// 账号名与昵称同为对外展示的身份标识，共用 PermNickname。
const (
	PermUsername = consts.PermNickname
	PermNickname = consts.PermNickname
	PermBio      = consts.PermSignature
)

// Field 一次资料更新中的单个字段：数据库列名 + 已清洗校验的值 + 所需权限位（0 表示不校验）。
type Field struct {
	Column string
	Value  any
	Perm   uint32
}

// store 只声明资料更新所需的最小 DAO 能力，避免 service 依赖整个 UserDAO。
type store interface {
	QueryActivePunishmentMask(userID uint32, now time.Time) (uint32, error)
	UpdateProfile(userID uint32, updates map[string]any) (*modeluser.User, error)
}

// Update 应用一组已清洗的字段更新。
// 先把所有字段所需的权限位合并成掩码，一次性查询生效中的惩罚掩码后统一校验，
// 避免逐字段查询数据库；再交给 DAO 落库并返回更新后的用户。
func Update(s store, userID uint32, fields []Field, now time.Time) (*modeluser.User, error) {
	if len(fields) == 0 {
		return nil, ErrNoFields
	}

	updates := make(map[string]any, len(fields))
	var required uint32
	for _, field := range fields {
		updates[field.Column] = field.Value
		required |= field.Perm
	}

	if required != 0 {
		mask, err := s.QueryActivePunishmentMask(userID, now)
		if err != nil {
			return nil, err
		}
		if mask&required != 0 {
			return nil, ErrPunished
		}
	}

	return s.UpdateProfile(userID, updates)
}

// NormalizeUsername 清洗并校验账号名：去除首尾空白，长度 2-64 字符，字符集受限。
func NormalizeUsername(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if !validLength(value, consts.MinUsernameLen, consts.MaxUsernameLen) {
		return "", false
	}
	if !usernamePattern.MatchString(value) {
		return "", false
	}
	return value, true
}

// NormalizeNickname 清洗并校验昵称：去除首尾空白，长度 1-64 字符。
// 允许中英文、emoji 等可见字符；拒绝控制字符与 Unicode 格式字符。
func NormalizeNickname(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if !validLength(value, consts.MinNicknameLen, consts.MaxNicknameLen) {
		return "", false
	}
	if hasInvisible(value, false) {
		return "", false
	}
	return value, true
}

// NormalizeBio 清洗并校验个性签名：统一换行符、去除首尾空白，最长 255 字符。
// 允许多行（前端使用多行输入框），其余控制字符与格式字符一律拒绝。
func NormalizeBio(raw string) (string, bool) {
	value := strings.ReplaceAll(raw, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > consts.MaxBioLen {
		return "", false
	}
	if hasInvisible(value, true) {
		return "", false
	}
	return value, true
}

// NormalizeGender 校验性别取值：0 未知 / 1 男 / 2 女 / 3 非二元性别。
func NormalizeGender(raw uint8) (uint8, bool) {
	if raw > consts.MaxGenderValue {
		return 0, false
	}
	return raw, true
}

// validLength 按「字符数」判断长度：MySQL utf8mb4 下 VARCHAR(n) 的 n 也是字符数，
// 用字节长度会把中文昵称误判为超长。
func validLength(value string, min int, max int) bool {
	length := utf8.RuneCountInString(value)
	return length >= min && length <= max
}

// hasInvisible 判断文本是否包含控制字符或 Unicode 格式字符（Cf，如零宽连接符 U+200D、
// 双向覆盖 U+202E）。这类字符肉眼不可见，却可用于伪造昵称或污染其它端展示。
// allowNewline 为 true 时放行换行符（个性签名允许多行）。
func hasInvisible(value string, allowNewline bool) bool {
	for _, r := range value {
		if allowNewline && r == '\n' {
			continue
		}
		if unicode.IsControl(r) || unicode.In(r, unicode.Cf) {
			return true
		}
	}
	return false
}
