package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/newfeed/community-news/services/media-service/internal/media/domain"
)

type Repository interface {
	Save(ctx context.Context, object domain.Object) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (Service) UploadContract(key string, ttl time.Duration) domain.PresignedURL {
	return domain.PresignedURL{URL: "/minio/" + key, Method: "PUT", ExpiresAt: time.Now().UTC().Add(ttl)}
}

func (s Service) SaveObject(ctx context.Context, ownerID, bucket, key, contentType string, size int64) (domain.Object, error) {
	object := domain.Object{
		ID:          newID(),
		OwnerID:     ownerID,
		Bucket:      bucket,
		Key:         key,
		ContentType: contentType,
		Size:        size,
		CreatedAt:   time.Now().UTC(),
	}
	return object, s.repo.Save(ctx, object)
}

func newID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return hex.EncodeToString(buf[0:4]) + "-" + hex.EncodeToString(buf[4:6]) + "-" + hex.EncodeToString(buf[6:8]) + "-" + hex.EncodeToString(buf[8:10]) + "-" + hex.EncodeToString(buf[10:])
}
