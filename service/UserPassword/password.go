// Package UserPassword 负责用户密码凭证的业务规则：
// 密码强度校验，以及把明文密码哈希后写入凭证表。
// 明文密码只在内存中短暂存在，任何日志/数据库都不应出现。
package UserPassword

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"miaoverse/consts"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/service/security/credentials/password"
)

// passwordWriter 只声明写入密码凭证所需的最小 DAO 能力，
// 避免 service 层依赖整个 UserDAO，便于测试替身注入。
type passwordWriter interface {
	UpsertPassword(credential *modeluser.UserCredential) error
}

// ValidateStrength 校验明文密码强度：
//   - 长度为 8-64 个字符（按字符数计算，中文/emoji 不会被字节长度误判）；
//   - 至少包含「字母 / 数字 / 符号」中的两类，拒绝纯数字、纯字母等弱口令；
//   - 不含控制字符、首尾无空白字符（避免用户输入不可见字符后无法再次登录）。
func ValidateStrength(plain string) bool {
	length := utf8.RuneCountInString(plain)
	if length < consts.MinPasswordLen || length > consts.MaxPasswordLen {
		return false
	}
	// 首尾空白通常是误输入（复制粘贴带入），直接拒绝而不是静默裁剪：
	// 静默裁剪会让用户以为空格也属于密码的一部分。
	if plain != strings.TrimSpace(plain) {
		return false
	}

	var hasLetter, hasDigit, hasSymbol bool
	for _, r := range plain {
		switch {
		case unicode.IsControl(r):
			return false
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			// 空格、标点、emoji 等一律计入「符号」类别（首尾空格已在上方拒绝）
			hasSymbol = true
		}
	}

	classes := 0
	for _, matched := range []bool{hasLetter, hasDigit, hasSymbol} {
		if matched {
			classes++
		}
	}
	return classes >= 2
}

// SetPassword 把明文密码 bcrypt 哈希后写入用户密码凭证，已存在则覆盖。
// 凭证结构复用 service/security/credentials/password，保证 credential_key 等字段与注册流程一致。
func SetPassword(writer passwordWriter, userID uint32, plain string) error {
	err, credentials := password.PasswordToCredentialStructure(userID, plain)
	if err != nil {
		return err
	}
	if len(credentials) == 0 || credentials[0] == nil {
		return errors.New("password credential build returned empty result")
	}
	return writer.UpsertPassword(credentials[0])
}
