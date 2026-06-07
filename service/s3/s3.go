package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type Config struct {
	Endpoint              string
	Region                string
	AccessKeyID           string
	SecretAccessKey       string
	SessionToken          string
	Bucket                string
	PublicBaseURL         string
	UsePathStyle          bool
	TempSignatureSecret   string
	TempSignatureDuration time.Duration
	TempLinkDuration      time.Duration
}

type Servant struct {
	client              *awss3.Client
	presigner           *awss3.PresignClient
	bucket              string
	publicBaseURL       string
	tempSignatureSecret []byte
	tempSignatureTTL    time.Duration
	tempLinkTTL         time.Duration
}

const (
	defaultTempSignatureDuration = 10 * time.Minute
	defaultTempLinkDuration      = 5 * time.Minute
	maxTempLinkDuration          = 7 * 24 * time.Hour
)

func NewServant(ctx context.Context, conf Config) (*Servant, error) {
	if strings.TrimSpace(conf.Region) == "" {
		return nil, errors.New("s3 region is required")
	}
	if strings.TrimSpace(conf.Bucket) == "" {
		return nil, errors.New("s3 bucket is required")
	}
	if strings.TrimSpace(conf.AccessKeyID) == "" {
		return nil, errors.New("s3 access_key_id is required")
	}
	if strings.TrimSpace(conf.SecretAccessKey) == "" {
		return nil, errors.New("s3 secret_access_key is required")
	}
	tempSignatureDuration, err := normalizeS3Duration(
		conf.TempSignatureDuration,
		defaultTempSignatureDuration,
		"s3 temp signature duration",
	)
	if err != nil {
		return nil, err
	}
	tempLinkDuration, err := normalizeS3Duration(
		conf.TempLinkDuration,
		defaultTempLinkDuration,
		"s3 temp link duration",
	)
	if err != nil {
		return nil, err
	}
	if tempLinkDuration > maxTempLinkDuration {
		return nil, fmt.Errorf("s3 temp link duration must not exceed %s", maxTempLinkDuration)
	}

	tempSignatureSecret := strings.TrimSpace(conf.TempSignatureSecret)
	if tempSignatureSecret == "" {
		tempSignatureSecret = conf.SecretAccessKey
	}

	awsConf, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(conf.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			conf.AccessKeyID,
			conf.SecretAccessKey,
			conf.SessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("load s3 config: %w", err)
	}

	client := awss3.NewFromConfig(awsConf, func(options *awss3.Options) {
		options.UsePathStyle = conf.UsePathStyle
		if strings.TrimSpace(conf.Endpoint) != "" {
			options.BaseEndpoint = aws.String(conf.Endpoint)
		}
	})

	return &Servant{
		client:              client,
		presigner:           awss3.NewPresignClient(client),
		bucket:              conf.Bucket,
		publicBaseURL:       strings.TrimRight(conf.PublicBaseURL, "/"),
		tempSignatureSecret: []byte(tempSignatureSecret),
		tempSignatureTTL:    tempSignatureDuration,
		tempLinkTTL:         tempLinkDuration,
	}, nil
}

func (s *Servant) PutObject(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	if err := validateObjectInput(key, body); err != nil {
		return "", err
	}

	input := &awss3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = aws.String(contentType)
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return "", fmt.Errorf("put s3 object %q: %w", key, err)
	}

	return s.ObjectURL(key), nil
}

func (s *Servant) GetObject(ctx context.Context, key string) (*awss3.GetObjectOutput, error) {
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("s3 object key is required")
	}

	output, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get s3 object %q: %w", key, err)
	}

	return output, nil
}

func (s *Servant) DeleteObject(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("s3 object key is required")
	}

	_, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete s3 object %q: %w", key, err)
	}

	return nil
}

func (s *Servant) ObjectExists(ctx context.Context, key string) (bool, error) {
	if strings.TrimSpace(key) == "" {
		return false, errors.New("s3 object key is required")
	}

	_, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NotFound", "NoSuchKey", "404":
			return false, nil
		}
	}

	return false, fmt.Errorf("head s3 object %q: %w", key, err)
}

func (s *Servant) PresignGetObject(ctx context.Context, key string, expires time.Duration) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("s3 object key is required")
	}

	request, err := s.presigner.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(options *awss3.PresignOptions) {
		options.Expires = expires
	})
	if err != nil {
		return "", fmt.Errorf("presign get s3 object %q: %w", key, err)
	}

	return request.URL, nil
}

func (s *Servant) PresignPutObject(ctx context.Context, key string, expires time.Duration, contentType string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("s3 object key is required")
	}

	input := &awss3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = aws.String(contentType)
	}

	request, err := s.presigner.PresignPutObject(ctx, input, func(options *awss3.PresignOptions) {
		options.Expires = expires
	})
	if err != nil {
		return "", fmt.Errorf("presign put s3 object %q: %w", key, err)
	}

	return request.URL, nil
}

func (s *Servant) ObjectURL(key string) string {
	if s.publicBaseURL == "" || strings.TrimSpace(key) == "" {
		return ""
	}

	escapedKey := (&url.URL{Path: strings.TrimLeft(key, "/")}).EscapedPath()
	return s.publicBaseURL + "/" + escapedKey
}

func (s *Servant) Bucket() string {
	return s.bucket
}

func (s *Servant) Client() *awss3.Client {
	return s.client
}

func validateObjectInput(key string, body io.Reader) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("s3 object key is required")
	}
	if body == nil {
		return errors.New("s3 object body is required")
	}
	return nil
}

func normalizeS3Duration(value time.Duration, defaultValue time.Duration, name string) (time.Duration, error) {
	if value == 0 {
		return defaultValue, nil
	}
	if value < 0 {
		return 0, fmt.Errorf("%s must be greater than 0", name)
	}
	return value, nil
}
