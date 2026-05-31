package domain

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

func (t RefreshToken) Expired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}

func (t RefreshToken) Active(now time.Time) bool {
	return t.RevokedAt == nil && !t.Expired(now)
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
