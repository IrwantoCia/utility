package s3

import "time"

// Provider represents an S3-compatible storage provider.
type Provider string

const (
	ProviderB2 Provider = "b2"
)

// Bucket represents an S3 bucket.
type Bucket struct {
	Name string
}

// Object represents an object in a bucket.
type Object struct {
	Key          string
	Size         int64
	LastModified time.Time
}
