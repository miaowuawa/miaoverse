package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"miaoverse/consts"
)

var (
	ErrTempSignatureInvalid = errors.New("s3 temp signature is invalid")
	ErrTempSignatureExpired = errors.New("s3 temp signature is expired")
)

type TempSignature struct {
	UID       string
	Signature string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type TempObjectLink struct {
	URL       string
	ExpiresAt time.Time
}

type tempSignaturePayload struct {
	UID       string
	Timestamp int64
}

func (s *Servant) CreateTempSignature(uid string) (*TempSignature, error) {
	return s.createTempSignature(uid, time.Now())
}

func (s *Servant) VerifyTempSignature(uid string, signature string) error {
	_, err := s.verifyTempSignature(uid, signature, time.Now())
	return err
}

func (s *Servant) GetTempObjectLink(ctx context.Context, uid string, signature string, key string) (*TempObjectLink, error) {
	now := time.Now()
	if _, err := s.verifyTempSignature(uid, signature, now); err != nil {
		return nil, err
	}

	link, err := s.PresignGetObject(ctx, key, s.tempLinkTTL)
	if err != nil {
		return nil, err
	}

	return &TempObjectLink{
		URL:       link,
		ExpiresAt: now.UTC().Add(s.tempLinkTTL),
	}, nil
}

func (s *Servant) createTempSignature(uid string, now time.Time) (*TempSignature, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return nil, errors.New("s3 temp signature uid is required")
	}
	if err := s.ensureTempSignatureConfig(); err != nil {
		return nil, err
	}

	issuedAt := now.UTC().Truncate(time.Second)
	payload, err := json.Marshal(tempSignaturePayload{
		UID:       uid,
		Timestamp: issuedAt.Unix(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal s3 temp signature payload: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signedValue := consts.TempSignatureVersion + "." + encodedPayload
	signature := base64.RawURLEncoding.EncodeToString(s.signTempPayload(signedValue))

	return &TempSignature{
		UID:       uid,
		Signature: signedValue + "." + signature,
		IssuedAt:  issuedAt,
		ExpiresAt: issuedAt.Add(s.tempSignatureTTL),
	}, nil
}

func (s *Servant) verifyTempSignature(uid string, signature string, now time.Time) (*tempSignaturePayload, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return nil, errors.New("s3 temp signature uid is required")
	}
	if err := s.ensureTempSignatureConfig(); err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(signature), ".")
	if len(parts) != 3 || parts[0] != consts.TempSignatureVersion {
		return nil, ErrTempSignatureInvalid
	}

	signedValue := parts[0] + "." + parts[1]
	gotSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: bad signature encoding", ErrTempSignatureInvalid)
	}
	wantSignature := s.signTempPayload(signedValue)
	if !hmac.Equal(gotSignature, wantSignature) {
		return nil, ErrTempSignatureInvalid
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: bad payload encoding", ErrTempSignatureInvalid)
	}

	var payload tempSignaturePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("%w: bad payload", ErrTempSignatureInvalid)
	}
	if payload.Timestamp <= 0 {
		return nil, fmt.Errorf("%w: bad timestamp", ErrTempSignatureInvalid)
	}
	if subtle.ConstantTimeCompare([]byte(payload.UID), []byte(uid)) != 1 {
		return nil, ErrTempSignatureInvalid
	}

	issuedAt := time.Unix(payload.Timestamp, 0).UTC()
	now = now.UTC()
	if issuedAt.After(now.Add(consts.TempSignatureClockGap)) {
		return nil, fmt.Errorf("%w: timestamp is in the future", ErrTempSignatureInvalid)
	}
	if now.After(issuedAt.Add(s.tempSignatureTTL)) {
		return nil, ErrTempSignatureExpired
	}

	return &payload, nil
}

func (s *Servant) ensureTempSignatureConfig() error {
	if s == nil {
		return errors.New("s3 servant is required")
	}
	if len(s.tempSignatureSecret) == 0 {
		return errors.New("s3 temp signature secret is required")
	}
	if s.tempSignatureTTL <= 0 {
		return errors.New("s3 temp signature duration must be greater than 0")
	}
	if s.tempLinkTTL <= 0 {
		return errors.New("s3 temp link duration must be greater than 0")
	}
	return nil
}

func (s *Servant) signTempPayload(value string) []byte {
	mac := hmac.New(sha256.New, s.tempSignatureSecret)
	mac.Write([]byte(value))
	return mac.Sum(nil)
}
