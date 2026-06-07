package s3

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCreateAndVerifyTempSignature(t *testing.T) {
	servant := newTestServant()
	now := time.Unix(1_700_000_000, 456_000_000).UTC()

	signature, err := servant.createTempSignature(" user-1 ", now)
	if err != nil {
		t.Fatal(err)
	}
	if signature.UID != "user-1" {
		t.Fatalf("UID = %q, want %q", signature.UID, "user-1")
	}
	if !signature.IssuedAt.Equal(time.Unix(1_700_000_000, 0).UTC()) {
		t.Fatalf("IssuedAt = %s, want timestamp truncated to second", signature.IssuedAt)
	}
	if !signature.ExpiresAt.Equal(signature.IssuedAt.Add(servant.tempSignatureTTL)) {
		t.Fatalf("ExpiresAt = %s, want %s", signature.ExpiresAt, signature.IssuedAt.Add(servant.tempSignatureTTL))
	}

	if _, err := servant.verifyTempSignature("user-1", signature.Signature, now.Add(time.Minute)); err != nil {
		t.Fatalf("verifyTempSignature() error = %v", err)
	}
}

func TestVerifyTempSignatureRejectsOtherUser(t *testing.T) {
	servant := newTestServant()
	now := time.Unix(1_700_000_000, 0).UTC()
	signature, err := servant.createTempSignature("user-1", now)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := servant.verifyTempSignature("user-2", signature.Signature, now); !errors.Is(err, ErrTempSignatureInvalid) {
		t.Fatalf("verifyTempSignature() error = %v, want ErrTempSignatureInvalid", err)
	}
}

func TestVerifyTempSignatureRejectsTamperedSignature(t *testing.T) {
	servant := newTestServant()
	now := time.Unix(1_700_000_000, 0).UTC()
	signature, err := servant.createTempSignature("user-1", now)
	if err != nil {
		t.Fatal(err)
	}
	tampered := signature.Signature[:len(signature.Signature)-1] + "A"
	if strings.HasSuffix(signature.Signature, "A") {
		tampered = signature.Signature[:len(signature.Signature)-1] + "B"
	}

	if _, err := servant.verifyTempSignature("user-1", tampered, now); !errors.Is(err, ErrTempSignatureInvalid) {
		t.Fatalf("verifyTempSignature() error = %v, want ErrTempSignatureInvalid", err)
	}
}

func TestVerifyTempSignatureRejectsExpiredSignature(t *testing.T) {
	servant := newTestServant()
	now := time.Unix(1_700_000_000, 0).UTC()
	signature, err := servant.createTempSignature("user-1", now)
	if err != nil {
		t.Fatal(err)
	}

	verifyAt := now.Add(servant.tempSignatureTTL).Add(time.Nanosecond)
	if _, err := servant.verifyTempSignature("user-1", signature.Signature, verifyAt); !errors.Is(err, ErrTempSignatureExpired) {
		t.Fatalf("verifyTempSignature() error = %v, want ErrTempSignatureExpired", err)
	}
}

func newTestServant() *Servant {
	return &Servant{
		tempSignatureSecret: []byte("test-secret"),
		tempSignatureTTL:    10 * time.Minute,
		tempLinkTTL:         5 * time.Minute,
	}
}
