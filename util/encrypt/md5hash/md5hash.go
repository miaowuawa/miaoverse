package md5hash

import (
	"crypto/md5"
	"encoding/hex"
)

func HashStr(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}
