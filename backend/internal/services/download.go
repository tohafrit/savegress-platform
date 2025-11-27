package services

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// DownloadService handles software downloads with S3 presigned URLs
type DownloadService struct {
	s3Client   *s3.Client
	presigner  *s3.PresignClient
	bucket     string
	keyPrefix  string
	urlExpiry  time.Duration
}

// DownloadConfig configures the download service
type DownloadConfig struct {
	// AWS Configuration
	Region          string
	Bucket          string
	KeyPrefix       string // Prefix for release files (e.g., "releases/")
	AccessKeyID     string
	SecretAccessKey string

	// URL settings
	URLExpiry time.Duration // How long signed URLs are valid (default: 1 hour)

	// Custom endpoint for S3-compatible storage
	Endpoint     string
	UsePathStyle bool
}

// ReleaseInfo contains information about a software release
type ReleaseInfo struct {
	Product   string   `json:"product"`
	Version   string   `json:"version"`
	Editions  []string `json:"editions"`
	Platforms []string `json:"platforms"`
}

// DownloadURL represents a presigned download URL
type DownloadURL struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"` // seconds
	Filename  string    `json:"filename"`
	Size      int64     `json:"size,omitempty"`
	Checksum  string    `json:"checksum,omitempty"`
}

// NewDownloadService creates a new download service
func NewDownloadService(ctx context.Context, cfg DownloadConfig) (*DownloadService, error) {
	var opts []func(*config.LoadOptions) error

	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		creds := credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)
		opts = append(opts, config.WithCredentialsProvider(creds))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// S3 client options
	s3Opts := []func(*s3.Options){}

	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	if cfg.UsePathStyle {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)
	presigner := s3.NewPresignClient(client)

	expiry := cfg.URLExpiry
	if expiry == 0 {
		expiry = 1 * time.Hour
	}

	return &DownloadService{
		s3Client:  client,
		presigner: presigner,
		bucket:    cfg.Bucket,
		keyPrefix: cfg.KeyPrefix,
		urlExpiry: expiry,
	}, nil
}

// GetDownloadURL generates a presigned URL for downloading a release
func (s *DownloadService) GetDownloadURL(ctx context.Context, product, version, edition, platform string) (*DownloadURL, error) {
	// Construct the S3 key for the release
	key := s.getReleaseKey(product, version, edition, platform)

	// Check if the object exists and get its metadata
	headOutput, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("release not found: %s/%s/%s/%s", product, version, edition, platform)
	}

	// Generate presigned URL
	presignedReq, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.urlExpiry
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	expiresAt := time.Now().Add(s.urlExpiry)

	result := &DownloadURL{
		URL:       presignedReq.URL,
		ExpiresAt: expiresAt,
		ExpiresIn: int64(s.urlExpiry.Seconds()),
		Filename:  s.getFilename(product, version, edition, platform),
	}

	if headOutput.ContentLength != nil {
		result.Size = *headOutput.ContentLength
	}

	// Get checksum if available from metadata
	if headOutput.Metadata != nil {
		if checksum, ok := headOutput.Metadata["sha256"]; ok {
			result.Checksum = checksum
		}
	}

	return result, nil
}

// ListReleases returns available releases
func (s *DownloadService) ListReleases(ctx context.Context) ([]ReleaseInfo, error) {
	// In a real implementation, this would query a database or S3 listing
	// For now, return hardcoded releases
	releases := []ReleaseInfo{
		{
			Product:   "cdc-engine",
			Version:   "1.0.0",
			Editions:  []string{"community", "pro", "enterprise"},
			Platforms: []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"},
		},
		{
			Product:   "cdc-broker",
			Version:   "1.0.0",
			Editions:  []string{"community", "pro", "enterprise"},
			Platforms: []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"},
		},
	}

	return releases, nil
}

// GetLatestVersion returns the latest version for a product
func (s *DownloadService) GetLatestVersion(ctx context.Context, product string) (string, error) {
	// In a real implementation, this would query S3 or a database
	switch product {
	case "cdc-engine", "cdc-broker":
		return "1.0.0", nil
	default:
		return "", fmt.Errorf("unknown product: %s", product)
	}
}

// getReleaseKey constructs the S3 key for a release file
func (s *DownloadService) getReleaseKey(product, version, edition, platform string) string {
	// Format: {prefix}/{product}/{version}/{edition}/{product}-{platform}
	// Example: releases/cdc-engine/1.0.0/pro/cdc-engine-linux-amd64
	filename := s.getFilename(product, version, edition, platform)
	return fmt.Sprintf("%s%s/%s/%s/%s", s.keyPrefix, product, version, edition, filename)
}

// getFilename constructs the filename for a release
func (s *DownloadService) getFilename(product, version, edition, platform string) string {
	ext := ""
	if platform == "windows-amd64" {
		ext = ".exe"
	}
	return fmt.Sprintf("%s-%s%s", product, platform, ext)
}

// ValidateDownloadRequest validates a download request
func (s *DownloadService) ValidateDownloadRequest(product, edition, platform string) error {
	validProducts := map[string]bool{
		"cdc-engine": true,
		"cdc-broker": true,
	}

	validEditions := map[string]bool{
		"community":  true,
		"pro":        true,
		"enterprise": true,
	}

	validPlatforms := map[string]bool{
		"linux-amd64":  true,
		"linux-arm64":  true,
		"darwin-amd64": true,
		"darwin-arm64": true,
		"windows-amd64": true,
	}

	if !validProducts[product] {
		return fmt.Errorf("invalid product: %s", product)
	}

	if !validEditions[edition] {
		return fmt.Errorf("invalid edition: %s", edition)
	}

	if !validPlatforms[platform] {
		return fmt.Errorf("invalid platform: %s", platform)
	}

	return nil
}

// CheckLicenseForEdition checks if user's license allows the requested edition
func (s *DownloadService) CheckLicenseForEdition(userTier, requestedEdition string) error {
	tierRanks := map[string]int{
		"community":  1,
		"pro":        2,
		"enterprise": 3,
	}

	userRank := tierRanks[userTier]
	requestedRank := tierRanks[requestedEdition]

	if userRank < requestedRank {
		return fmt.Errorf("license tier %s does not include %s edition", userTier, requestedEdition)
	}

	return nil
}

// UploadRelease uploads a new release to S3 (for admin use)
func (s *DownloadService) UploadRelease(ctx context.Context, product, version, edition, platform string, data []byte, checksum string) error {
	key := s.getReleaseKey(product, version, edition, platform)

	metadata := map[string]string{}
	if checksum != "" {
		metadata["sha256"] = checksum
	}

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		Body:     bytes.NewReader(data),
		Metadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to upload release: %w", err)
	}

	return nil
}
