package bcrypthash

import "golang.org/x/crypto/bcrypt"

// 生成密码哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12) // 12是成本因子，越高越安全
	return string(bytes), err
}

// 验证密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
