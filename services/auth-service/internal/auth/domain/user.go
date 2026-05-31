package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrRefreshNotFound    = errors.New("refresh token not found")
)

type Role string

const (
	RoleReader Role = "reader"
	RoleEditor Role = "editor"
	RoleAdmin  Role = "admin"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         Role
	Verified     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(id, email, passwordHash, firstName, lastName string) (*User, error) {
	now := time.Now().UTC()
	user := &User{
		ID:           id,
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: passwordHash,
		FirstName:    strings.TrimSpace(firstName),
		LastName:     strings.TrimSpace(lastName),
		Role:         RoleReader,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return user, user.Validate()
}

func (u *User) Validate() error {
	if !strings.Contains(u.Email, "@") {
		return ErrInvalidEmail
	}
	if u.PasswordHash == "" {
		return ErrInvalidPassword
	}
	return nil
}
