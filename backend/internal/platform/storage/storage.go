// Package storage issues presigned upload and download URLs. Files never pass through the API.
// The fake provider (default) returns local URLs and accepts uploads into memory for development.
package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Port is what modules depend on.
type Port interface {
	PresignPut(ctx context.Context, key, contentType string, private bool, ttl time.Duration) (string, error)
	PresignGet(ctx context.Context, key string, private bool, ttl time.Duration) (string, error)
	PublicURL(key string) string
}

// Fake returns predictable local URLs (no external calls).
type Fake struct{ BaseURL string }

func (f Fake) PresignPut(_ context.Context, key, _ string, _ bool, ttl time.Duration) (string, error) {
	return fmt.Sprintf("%s/dev/uploads/%s?expires=%d", f.BaseURL, url.PathEscape(key), time.Now().Add(ttl).Unix()), nil
}
func (f Fake) PresignGet(_ context.Context, key string, _ bool, ttl time.Duration) (string, error) {
	return fmt.Sprintf("%s/dev/files/%s?expires=%d", f.BaseURL, url.PathEscape(key), time.Now().Add(ttl).Unix()), nil
}
func (f Fake) PublicURL(key string) string { return f.BaseURL + "/dev/files/" + url.PathEscape(key) }

// S3 uses any S3-compatible store (MinIO locally, AWS S3 in production).
type S3 struct {
	presign       *s3.PresignClient
	publicBucket  string
	privateBucket string
	publicBaseURL string
}

// S3Config configures the S3 adapter.
type S3Config struct {
	Endpoint, Region, PublicBucket, PrivateBucket, AccessKeyID, SecretAccessKey, PublicBaseURL string
	ForcePathStyle                                                                             bool
}

// NewS3 builds the S3 adapter.
func NewS3(cfg S3Config) *S3 {
	client := s3.New(s3.Options{
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: cfg.ForcePathStyle,
	})
	return &S3{presign: s3.NewPresignClient(client), publicBucket: cfg.PublicBucket, privateBucket: cfg.PrivateBucket, publicBaseURL: cfg.PublicBaseURL}
}

func (s *S3) bucket(private bool) string {
	if private {
		return s.privateBucket
	}
	return s.publicBucket
}

func (s *S3) PresignPut(ctx context.Context, key, contentType string, private bool, ttl time.Duration) (string, error) {
	r, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket(private)), Key: aws.String(key), ContentType: aws.String(contentType)},
		s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return r.URL, nil
}

func (s *S3) PresignGet(ctx context.Context, key string, private bool, ttl time.Duration) (string, error) {
	r, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket(private)), Key: aws.String(key)}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return r.URL, nil
}

func (s *S3) PublicURL(key string) string { return s.publicBaseURL + "/" + key }
