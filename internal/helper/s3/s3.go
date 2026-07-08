// Package s3 provides an S3-compatible storage client for interacting with
// object storage services like Backblaze B2, AWS S3, Cloudflare R2, and MinIO.
//
// # Quick Start
//
//	client, err := s3.New(s3.Config{
//	    Provider:  s3.ProviderB2,
//	    AccessKey: "your-key-id",
//	    SecretKey: "your-app-key",
//	})
//	buckets, _ := client.ListBuckets(ctx)
//
// # Environment Variables
//
// Configuration falls back to environment variables when fields are empty:
//
//	S3_PROVIDER   - provider type ("b2", "aws", etc.)
//	S3_ENDPOINT   - custom endpoint (overrides provider default)
//	S3_ACCESS_KEY - access key ID
//	S3_SECRET_KEY - secret access key
//	S3_SECURE     - set to "false" or "0" to disable HTTPS
//
// # Adding Providers
//
// Add new providers to the providerEndpoints map in provider.go:
//
//	var providerEndpoints = map[Provider]string{
//	    ProviderB2: "s3.us-west-004.backblazeb2.com",
//	}
package s3

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 is an S3-compatible storage client.
type S3 struct {
	client *minio.Client
}

// New creates an S3 client.
func New(cfg Config) (*S3, error) {
	applyEnvOverrides(&cfg)
	endpoint := resolveEndpoint(cfg)
	if endpoint == "" {
		return nil, ErrNoEndpoint
	}

	secure := true
	if !cfg.Secure {
		secure = false
	}
	// Allow explicit override via environment variable.
	secureEnv := os.Getenv("S3_SECURE")
	if secureEnv == "false" || secureEnv == "0" {
		secure = false
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	return &S3{client: client}, nil
}

// ListBuckets returns all buckets.
func (s *S3) ListBuckets(ctx context.Context) ([]Bucket, error) {
	minioBuckets, err := s.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrListBuckets, err)
	}
	buckets := make([]Bucket, len(minioBuckets))
	for i, b := range minioBuckets {
		buckets[i] = Bucket{Name: b.Name}
	}
	return buckets, nil
}

// ListObjects returns objects in the given bucket.
// If prefix is non-empty, only objects matching the prefix are returned.
func (s *S3) ListObjects(ctx context.Context, bucket, prefix string) ([]Object, error) {
	var objects []Object
	for obj := range s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("%w: %w", ErrListObjects, obj.Err)
		}
		objects = append(objects, Object{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
		})
	}
	return objects, nil
}

// UploadFile uploads a local file to S3.
// The progress channel receives cumulative bytes uploaded per 1MB chunk;
// when upload completes, the caller should detect completion by consuming
// all progress events and observing the function returns nil.
func (s *S3) UploadFile(ctx context.Context, bucket, key, filePath string, progress chan<- int64) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUpload, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUploadStat, err)
	}

	reader := newProgressReader(file, stat.Size(), progress)

	_, err = s.client.PutObject(ctx, bucket, key, reader, stat.Size(), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUpload, err)
	}

	return nil
}

// progressReader wraps an io.Reader and reports progress via channel.
type progressReader struct {
	reader   io.Reader
	total    int64
	sent     int64
	progress chan<- int64
}

func newProgressReader(r io.Reader, total int64, progress chan<- int64) *progressReader {
	return &progressReader{reader: r, total: total, progress: progress}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.sent += int64(n)
	if pr.progress != nil {
		select {
		case pr.progress <- pr.sent:
		default:
		}
	}
	return n, err
}
