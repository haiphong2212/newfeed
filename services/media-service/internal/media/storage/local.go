package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	root    string
	baseURL string
}

func NewLocalStorage(root, baseURL string) *LocalStorage {
	return &LocalStorage{root: root, baseURL: baseURL}
}

func (s *LocalStorage) Put(_ context.Context, bucket, key, _ string, body io.Reader) (StoredObject, error) {
	targetDir := filepath.Join(s.root, bucket, filepath.Dir(key))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return StoredObject{}, err
	}
	targetPath := filepath.Join(s.root, bucket, key)
	target, err := os.Create(targetPath)
	if err != nil {
		return StoredObject{}, err
	}
	size, copyErr := io.Copy(target, body)
	closeErr := target.Close()
	if copyErr != nil {
		return StoredObject{}, copyErr
	}
	if closeErr != nil {
		return StoredObject{}, closeErr
	}
	return StoredObject{Bucket: bucket, Key: key, Size: size, URL: s.PublicURL(bucket, key)}, nil
}

func (s *LocalStorage) PublicURL(bucket, key string) string {
	if s.baseURL == "" {
		return ""
	}
	return s.baseURL + "/objects/" + bucket + "/" + key
}
