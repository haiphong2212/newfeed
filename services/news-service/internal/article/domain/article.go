package domain

import (
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusReview    Status = "review"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

var ErrInvalidArticle = errors.New("invalid article")

type Article struct {
	ID        string
	AuthorID  string
	Title     string
	Slug      string
	Content   string
	Category  string
	Tags      []string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a Article) DiscussionRoomName() string {
	return strings.TrimSpace(a.Title) + " Live Discussion"
}

func (a Article) ValidateForPublish() error {
	if a.ID == "" || a.AuthorID == "" || a.Title == "" || a.Content == "" || a.Category == "" {
		return ErrInvalidArticle
	}
	return nil
}
