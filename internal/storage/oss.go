package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// OSSStorage implements Storage using an S3-compatible object store (Aliyun OSS, MinIO, AWS S3, etc.).
type OSSStorage struct {
	client *s3.Client
	bucket string
}

// OSSConfig holds the configuration for S3-compatible storage.
type OSSConfig struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
}

// NewOSS creates a Storage backed by an S3-compatible service.
func NewOSS(cfg OSSConfig) (*OSSStorage, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("OSS_ENDPOINT, OSS_BUCKET, OSS_ACCESS_KEY, and OSS_SECRET_KEY are required when STORAGE_DRIVER=oss")
	}
	region := cfg.Region
	if region == "" {
		region = "us-east-1" // default region for S3-compatible services
	}

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.Endpoint),
		Region:       region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		UsePathStyle: true, // Required for most S3-compatible services (OSS, MinIO)
	})

	return &OSSStorage{client: client, bucket: cfg.Bucket}, nil
}

// PresignPut generates a presigned PUT URL.
func (s *OSSStorage) PresignPut(ctx context.Context, key string, contentType string, expires time.Duration) (string, http.Header, error) {
	presignClient := s3.NewPresignClient(s.client)
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	result, err := presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(expires))
	if err != nil {
		return "", nil, fmt.Errorf("oss presign put: %w", err)
	}

	h := http.Header{}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	return result.URL, h, nil
}

// PresignGet generates a presigned GET URL.
func (s *OSSStorage) PresignGet(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	result, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("oss presign get: %w", err)
	}
	return result.URL, nil
}

// GetObject downloads an object from storage.
func (s *OSSStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("oss get object: %w", err)
	}
	return output.Body, nil
}

// DeleteObject removes an object from storage.
func (s *OSSStorage) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("oss delete object: %w", err)
	}
	return nil
}
