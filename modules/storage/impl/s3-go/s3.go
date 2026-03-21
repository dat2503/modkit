// Package s3 implements the StorageService interface using AWS S3 (and S3-compatible providers).
package s3

import (
	"context"
	"io"
	"time"

	contracts "github.com/dat2503/modkit/contracts/go"
)

// Config holds the configuration for the S3 storage provider.
type Config struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	// Endpoint is the custom endpoint for S3-compatible providers (e.g. Cloudflare R2, MinIO).
	// Leave empty for AWS S3.
	Endpoint      string
	PublicBaseURL string
}

// Service implements contracts.StorageService using AWS S3.
type Service struct {
	cfg Config
	// TODO: add aws-sdk-go-v2 s3 client
}

// New creates a new S3 storage service.
func New(cfg Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Upload(ctx context.Context, key string, r io.Reader, opts contracts.UploadOptions) (*contracts.UploadResult, error) {
	// TODO: implement using aws-sdk-go-v2 s3.PutObject
	panic("not implemented")
}

func (s *Service) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// TODO: implement using aws-sdk-go-v2 s3.GetObject
	panic("not implemented")
}

func (s *Service) Delete(ctx context.Context, key string) error {
	// TODO: implement using aws-sdk-go-v2 s3.DeleteObject
	panic("not implemented")
}

func (s *Service) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// TODO: implement using aws-sdk-go-v2 s3.PresignGetObject
	panic("not implemented")
}

func (s *Service) PublicURL(ctx context.Context, key string) (string, error) {
	// TODO: return s.cfg.PublicBaseURL + "/" + key if PublicBaseURL is set
	panic("not implemented")
}

func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	// TODO: implement using aws-sdk-go-v2 s3.HeadObject
	panic("not implemented")
}
