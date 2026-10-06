package UserPassword

import (
	"errors"
	"strings"
	"testing"

	modeluser "miaoverse/model/dao/user"
	"miaoverse/util"
)

type fakeWriter struct {
	got      *modeluser.UserCredential
	err      error
	callTime int
}

func (f *fakeWriter) UpsertPassword(credential *modeluser.UserCredential) error {
	f.callTime++
	f.got = credential
	return f.err
}

func TestValidateStrength(t *testing.T) {
	cases := []struct {
		name     string
		password string
		ok       bool
	}{
		{"letters and digits", "miaoverse2026", true},
		{"letters and symbols", "miaoverse!@#", true},
		{"digits and symbols", "12345678!!", true},
		{"passphrase with inner space", "miaoverse 2026 pass", true},
		{"chinese and digits", "喵星人密码2026", true},
		{"too short", "abc123!", false},
		{"letters only", "abcdefghij", false},
		{"digits only", "1234567890", false},
		{"symbols only", "!!!!!!!!!!", false},
		{"leading space", " miaoverse2026", false},
		{"trailing space", "miaoverse2026 ", false},
		{"tab inside", "miaoverse\t2026", false},
		{"empty", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ValidateStrength(c.password); got != c.ok {
				t.Fatalf("ValidateStrength(%q) = %v, want %v", c.password, got, c.ok)
			}
		})
	}
}

func TestValidateStrengthRejectsOverlongPassword(t *testing.T) {
	// bcrypt 只取前 72 字节，超过上限的部分会被静默忽略，因此必须显式拒绝
	password := strings.Repeat("a1", 40) // 80 字符
	if ValidateStrength(password) {
		t.Fatalf("ValidateStrength should reject %d-character password", len(password))
	}
}

func TestSetPasswordStoresBcryptHashNotPlaintext(t *testing.T) {
	writer := &fakeWriter{}
	plain := "miaoverse2026"

	if err := SetPassword(writer, 10001, plain); err != nil {
		t.Fatalf("SetPassword error = %v", err)
	}
	if writer.got == nil {
		t.Fatal("SetPassword did not write a credential")
	}
	if writer.got.UserID != 10001 {
		t.Fatalf("credential user_id = %d, want 10001", writer.got.UserID)
	}
	if writer.got.CredentialValue == plain {
		t.Fatal("credential value must not be the plaintext password")
	}
	if strings.Contains(writer.got.CredentialValue, plain) {
		t.Fatal("credential value must not contain the plaintext password")
	}
	if !util.Security.CheckPassword(plain, writer.got.CredentialValue) {
		t.Fatal("stored hash does not verify against the plaintext password")
	}
	if util.Security.CheckPassword("wrong-password", writer.got.CredentialValue) {
		t.Fatal("stored hash must not verify a wrong password")
	}
}

func TestSetPasswordOverwritesExistingCredential(t *testing.T) {
	writer := &fakeWriter{}
	if err := SetPassword(writer, 10001, "first-password1"); err != nil {
		t.Fatalf("SetPassword error = %v", err)
	}
	first := writer.got.CredentialValue

	if err := SetPassword(writer, 10001, "second-password2"); err != nil {
		t.Fatalf("SetPassword error = %v", err)
	}
	second := writer.got.CredentialValue

	if first == second {
		t.Fatal("bcrypt hash should differ between two different passwords")
	}
	if !util.Security.CheckPassword("second-password2", second) {
		t.Fatal("latest hash should verify the latest password")
	}
	if util.Security.CheckPassword("first-password1", second) {
		t.Fatal("latest hash must not verify the previous password")
	}
	if writer.callTime != 2 {
		t.Fatalf("UpsertPassword called %d times, want 2", writer.callTime)
	}
}

func TestSetPasswordPropagatesDAOError(t *testing.T) {
	writer := &fakeWriter{err: errors.New("db down")}
	if err := SetPassword(writer, 10001, "miaoverse2026"); err == nil {
		t.Fatal("SetPassword should propagate DAO error")
	}
}

func TestSetPasswordKeepsCredentialKeyStable(t *testing.T) {
	// credential_key 必须与注册流程一致，否则 UserCheck.PasswordSet / QueryCredential 会查不到
	writer := &fakeWriter{}
	if err := SetPassword(writer, 10001, "miaoverse2026"); err != nil {
		t.Fatalf("SetPassword error = %v", err)
	}
	if writer.got.CredentialKey != "bcrypt" {
		t.Fatalf("credential_key = %q, want %q", writer.got.CredentialKey, "bcrypt")
	}
}
