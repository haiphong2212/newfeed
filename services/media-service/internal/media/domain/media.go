package domain

import "time"

type Object struct {
	ID          string
	OwnerID     string
	Bucket      string
	Key         string
	ContentType string
	Size        int64
	CreatedAt   time.Time
}

type PresignedURL struct {
	URL       string
	Method    string
	ExpiresAt time.Time
}
