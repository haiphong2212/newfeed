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

type Comment struct {
	ID        string     `json:"id"`
	ArticleID string     `json:"article_id"`
	UserID    string     `json:"user_id"`
	ParentID  string     `json:"parent_id,omitempty"`
	Body      string     `json:"body"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Share struct {
	ID         string    `json:"id"`
	ArticleID  string    `json:"article_id"`
	UserID     string    `json:"user_id"`
	Caption    string    `json:"caption"`
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"created_at"`
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
