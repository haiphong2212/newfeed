package domain

import "time"

type Room struct {
	ID         string     `json:"id"`
	ArticleID  string     `json:"article_id"`
	Name       string     `json:"name"`
	Locked     bool       `json:"locked"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Message struct {
	ID          string         `json:"id"`
	RoomID      string         `json:"room_id"`
	UserID      string         `json:"user_id"`
	MessageType string         `json:"message_type"`
	Body        string         `json:"body"`
	Metadata    map[string]any `json:"metadata"`
	EditedAt    *time.Time     `json:"edited_at,omitempty"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Presence struct {
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Online    bool      `json:"online"`
	Typing    bool      `json:"typing"`
	UpdatedAt time.Time `json:"updated_at"`
}
