// Package s3 implements the StorageService interface using AWS S3 (and S3-compatible providers).
package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

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
	cfg    Config
	client *s3.Client
}

// New creates a new S3 storage service.
func New(cfg Config) (*Service, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3: Bucket is required")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("s3: Region is required")
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("s3: load config: %w", err)
	}

	s3opts := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		s3opts = append(s3opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3opts...)
	return &Service{cfg: cfg, client: client}, nil
}

func (s *Service) Upload(ctx context.Context, key string, r io.Reader, opts contracts.UploadOptions) (*contracts.UploadResult, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
		Body:   r,
	}
	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}
	if opts.Public {
		input.ACL = types.ObjectCannedACLPublicRead
	}
	if len(opts.Metadata) > 0 {
		input.Metadata = opts.Metadata
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("s3: upload %q: %w", key, err)
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.Bucket, s.cfg.Region, key)
	if s.cfg.Endpoint != "" {
		url = fmt.Sprintf("%s/%s/%s", s.cfg.Endpoint, s.cfg.Bucket, key)
	}
	return &contracts.UploadResult{Key: key, URL: url}, nil
}

func (s *Service) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3: download %q: %w", key, err)
	}
	return out.Body, nil
}

func (s *Service) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3: delete %q: %w", key, err)
	}
	return nil
}

func (s *Service) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	result, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("s3: sign url %q: %w", key, err)
	}
	return result.URL, nil
}

func (s *Service) PublicURL(ctx context.Context, key string) (string, error) {
	if s.cfg.PublicBaseURL == "" {
		return "", fmt.Errorf("s3: PublicBaseURL is not configured")
	}
	return s.cfg.PublicBaseURL + "/" + key, nil
}

func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			code := apiErr.ErrorCode()
			if code == "NotFound" || code == "NoSuchKey" || code == "404" {
				return false, nil
			}
		}
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, fmt.Errorf("s3: exists %q: %w", key, err)
	}
	return true, nil
}
