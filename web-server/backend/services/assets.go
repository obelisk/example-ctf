package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/obelisk/example-ctf/config"
	log "github.com/sirupsen/logrus"
)

type cachedAsset struct {
	url       string
	expiresAt time.Time
}

// AssetService is used to manage access to static assets stored in an S3 bucket
type AssetService struct {
	presignerClient *s3.PresignClient
	awsCfg          awsConfig.Config
	cfg             *config.Config
	cache           map[string]*cachedAsset
	mutex           sync.RWMutex
}

// NewAssetService initializes the AssetService with the configured AWS credentials
func NewAssetService(cfg *config.Config) (*AssetService, error) {
	// Load AWS config (env vars, shared config file, EC2/ECS metadata, etc.)
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	// Create an S3 client…
	client := s3.NewFromConfig(awsCfg)
	// …and wrap it in a presigner
	presigner := s3.NewPresignClient(client)

	return &AssetService{
		presignerClient: presigner,
		awsCfg:          awsCfg,
		cfg:             cfg,
		cache:           make(map[string]*cachedAsset),
	}, nil
}

// GetAsset will return an S3 presigned URL for the requested asset.
// The presigned URLs are cached in memory.
func (s *AssetService) GetAsset(ctx context.Context, path string) (string, error) {
	// Check cache first
	s.mutex.RLock()
	if cached, exists := s.cache[path]; exists && time.Now().Before(cached.expiresAt) {
		s.mutex.RUnlock()
		log.WithFields(log.Fields{
			"path":      path,
			"cached":    true,
			"expiresAt": cached.expiresAt,
		}).Debug("Asset URL served from cache")
		return cached.url, nil
	}
	s.mutex.RUnlock()

	// Cache miss or expired - generate new presigned URL
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Double-check cache after acquiring write lock
	if cached, exists := s.cache[path]; exists && time.Now().Before(cached.expiresAt) {
		log.WithFields(log.Fields{
			"path":      path,
			"cached":    true,
			"expiresAt": cached.expiresAt,
		}).Debug("Asset URL served from cache (double-check)")
		return cached.url, nil
	}

	// Generate new presigned URL
	presignResult, err := s.presignerClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.cfg.AwsConfig.BucketName),
			Key:    aws.String(path),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = 60 * time.Minute
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to PresignGetObject: %v", err)
	}

	// Cache the presigned URL
	expiresAt := time.Now().Add(45 * time.Minute)
	s.cache[path] = &cachedAsset{
		url:       presignResult.URL,
		expiresAt: expiresAt,
	}

	log.WithFields(log.Fields{
		"path":      path,
		"cached":    false,
		"expiresAt": expiresAt,
	}).Debug("Asset URL generated and cached")

	return presignResult.URL, nil
}
