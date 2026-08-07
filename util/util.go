package util

import (
	validator "github.com/go-playground/validator/v10"
	buildstringutil "miaoverse/util/buildstring"
	md5hashutil "miaoverse/util/encrypt/md5hash"
	filetypeutil "miaoverse/util/filetype"
	mathsutil "miaoverse/util/maths"
	securityutil "miaoverse/util/security"
	bcrypthashutil "miaoverse/util/security/bcrypthash"
	validateutil "miaoverse/util/validate"
)

var (
	BuildString buildStringNamespace
	FileType    fileTypeNamespace
	Maths       mathsNamespace
	MD5Hash     md5HashNamespace
	Security    securityNamespace
	Validate    validateNamespace
)

type buildStringNamespace struct{}

func (buildStringNamespace) GenerateUsername() (string, error) {
	return buildstringutil.GenerateUsername()
}

type fileTypeNamespace struct{}

func (fileTypeNamespace) Normalize(value string, mimeType string) uint8 {
	return filetypeutil.Normalize(value, mimeType)
}

type mathsNamespace struct{}

func (mathsNamespace) RandomIntLimited(min int, max int) (int, error) {
	return mathsutil.RandomIntLimited(min, max)
}

type md5HashNamespace struct{}

func (md5HashNamespace) HashStr(value string) string {
	return md5hashutil.HashStr(value)
}

type securityNamespace struct{}

func (securityNamespace) ValidateAvalue(a string) (bool, error) {
	return securityutil.ValidateAvalue(a)
}

func (securityNamespace) HashPassword(password string) (string, error) {
	return bcrypthashutil.HashPassword(password)
}

func (securityNamespace) CheckPassword(password string, hash string) bool {
	return bcrypthashutil.CheckPassword(password, hash)
}

type validateNamespace struct{}

func (validateNamespace) InitialValidator(v *validator.Validate) error {
	return validateutil.InitialValidator(v)
}

func (validateNamespace) ParseUA(ua string) string {
	return validateutil.ParseUA(ua)
}
