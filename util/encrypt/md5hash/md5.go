package md5hash

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

// HashStr 对字符串进行MD5哈希（基础版，无盐）
func HashStr(str string) string {
	return HashBytes([]byte(str))
}

// HashBytes 对字节切片进行MD5哈希（通用版）
func HashBytes(data []byte) string {
	h := md5.New()
	// 检查Write返回值（规范写法，避免潜在问题）
	n, err := h.Write(data)
	if err != nil || n != len(data) {
		return "" // 写入失败时返回空字符串，也可根据需求返回error
	}
	return hex.EncodeToString(h.Sum(nil))
}

// HashStrWithSalt 带盐的MD5哈希（提升安全性，适合敏感数据）
func HashStrWithSalt(str, salt string) string {
	// 盐值拼接方式：str + salt（也可 salt + str 或更复杂的拼接）
	return HashBytes([]byte(str + salt))
}

// HashReader 对io.Reader（如文件）进行MD5哈希（适合大文件）
func HashReader(r io.Reader) (string, error) {
	h := md5.New()
	// 直接从Reader读取数据，避免加载整个文件到内存
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
