package s3

import "errors"

// Sentinel errors for the s3 package.
var (
	ErrNoEndpoint   = errors.New("s3: endpoint is required, use Provider or set Endpoint/ENV")
	ErrClientCreate = errors.New("s3: failed to create client")
	ErrListBuckets  = errors.New("s3: list buckets failed")
	ErrListObjects  = errors.New("s3: list objects failed")
	ErrUpload       = errors.New("s3: upload failed")
	ErrUploadStat   = errors.New("s3: upload stat failed")
)
