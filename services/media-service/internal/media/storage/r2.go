package storage

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	PublicURL       string
}

type R2Storage struct {
	client    *s3.Client
	publicURL string
}

func NewR2Storage(cfg R2Config) *R2Storage {
	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(cfg.Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		UsePathStyle: true,
	})
	return &R2Storage{client: client, publicURL: strings.TrimRight(cfg.PublicURL, "/")}
}

func (s *R2Storage) Put(ctx context.Context, bucket, key, contentType string, body io.Reader) (StoredObject, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return StoredObject{}, err
	}
	return StoredObject{Bucket: bucket, Key: key, URL: s.PublicURL(bucket, key)}, nil
}

func (s *R2Storage) PublicURL(bucket, key string) string {
	if s.publicURL == "" {
		return ""
	}
	return s.publicURL + "/" + bucket + "/" + key
}
