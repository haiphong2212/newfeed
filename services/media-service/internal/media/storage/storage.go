package storage

import (
	"context"
	"io"
)

type StoredObject struct {
	Bucket string
	Key    string
	Size   int64
	URL    string
}

type ObjectStorage interface {
	Put(ctx context.Context, bucket, key, contentType string, body io.Reader) (StoredObject, error)
	PublicURL(bucket, key string) string
}
