package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// StorageService handles S3/MinIO storage operations
type StorageService struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucket     string
	region     string
	endpoint   string // For MinIO compatibility
	urlExpiry  time.Duration
}

// StorageConfig holds configuration for the storage service
type StorageConfig struct {
	Bucket          string
	Region          string
	Endpoint        string // Optional: for MinIO or S3-compatible storage
	AccessKeyID     string
	SecretAccessKey string
	URLExpiry       time.Duration // Default: 1 hour
}

// NewStorageService creates a new storage service
func NewStorageService(cfg StorageConfig) (*StorageService, error) {
	if cfg.URLExpiry == 0 {
		cfg.URLExpiry = time.Hour
	}

	// Build AWS config options
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	// Add credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with optional custom endpoint (for MinIO)
	var client *s3.Client
	if cfg.Endpoint != "" {
		client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // Required for MinIO
		})
	} else {
		client = s3.NewFromConfig(awsCfg)
	}

	return &StorageService{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.Bucket,
		region:    cfg.Region,
		endpoint:  cfg.Endpoint,
		urlExpiry: cfg.URLExpiry,
	}, nil
}

// DownloadInfo contains information about a downloadable file
type DownloadInfo struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"` // seconds
	Filename  string    `json:"filename"`
	Size      int64     `json:"size,omitempty"`
}

// GenerateDownloadURL generates a pre-signed URL for downloading a file
func (s *StorageService) GenerateDownloadURL(ctx context.Context, key string) (*DownloadInfo, error) {
	// Get object metadata to verify it exists and get size
	headOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Generate pre-signed URL
	presignedReq, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.urlExpiry))
	if err != nil {
		return nil, fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	expiresAt := time.Now().Add(s.urlExpiry)
	return &DownloadInfo{
		URL:       presignedReq.URL,
		ExpiresAt: expiresAt,
		ExpiresIn: int64(s.urlExpiry.Seconds()),
		Filename:  key,
		Size:      aws.ToInt64(headOutput.ContentLength),
	}, nil
}

// GenerateUploadURL generates a pre-signed URL for uploading a file
func (s *StorageService) GenerateUploadURL(ctx context.Context, key, contentType string) (*DownloadInfo, error) {
	presignedReq, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(s.urlExpiry))
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	expiresAt := time.Now().Add(s.urlExpiry)
	return &DownloadInfo{
		URL:       presignedReq.URL,
		ExpiresAt: expiresAt,
		ExpiresIn: int64(s.urlExpiry.Seconds()),
		Filename:  key,
	}, nil
}

// ListFiles lists files in a prefix (directory)
func (s *StorageService) ListFiles(ctx context.Context, prefix string) ([]FileInfo, error) {
	output, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]FileInfo, 0, len(output.Contents))
	for _, obj := range output.Contents {
		files = append(files, FileInfo{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: aws.ToTime(obj.LastModified),
			ETag:         aws.ToString(obj.ETag),
		})
	}
	return files, nil
}

// FileInfo contains metadata about a stored file
type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag"`
}

// DeleteFile deletes a file from storage
func (s *StorageService) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// FileExists checks if a file exists in storage
func (s *StorageService) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil // File doesn't exist
	}
	return true, nil
}

// GetReleaseKey constructs the S3 key for a release binary
func GetReleaseKey(product, version, platform, edition string) string {
	// Format: releases/{product}/{version}/{edition}/{product}-{platform}
	// Example: releases/cdc-engine/1.0.0/pro/cdc-engine-linux-amd64
	filename := fmt.Sprintf("%s-%s", product, platform)
	if platform == "windows-amd64" {
		filename += ".exe"
	}
	return fmt.Sprintf("releases/%s/%s/%s/%s", product, version, edition, filename)
}
