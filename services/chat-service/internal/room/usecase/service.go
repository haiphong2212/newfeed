package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/newfeed/community-news/services/chat-service/internal/room/domain"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrRoomLocked   = errors.New("room is locked")
	ErrRoomArchived = errors.New("room is archived")
)

type Repository interface {
	CreateRoom(ctx context.Context, room domain.Room) (domain.Room, error)
	GetRoom(ctx context.Context, id string) (domain.Room, error)
	GetRoomByArticle(ctx context.Context, articleID string) (domain.Room, error)
	ListRooms(ctx context.Context, limit int, cursor time.Time) ([]domain.Room, error)
	SetRoomLocked(ctx context.Context, id string, locked bool) error
	ArchiveRoom(ctx context.Context, id string) error
	CreateMessage(ctx context.Context, message domain.Message) (domain.Message, error)
	ListMessages(ctx context.Context, roomID string, limit int, cursor time.Time) ([]domain.Message, error)
	EditMessage(ctx context.Context, messageID, userID, body string) (domain.Message, error)
	DeleteMessage(ctx context.Context, messageID, userID string) error
	UpsertPresence(ctx context.Context, presence domain.Presence) error
	ListPresence(ctx context.Context, roomID string) ([]domain.Presence, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) RoomForArticle(articleID, articleTitle string) domain.Room {
	return domain.Room{ID: articleID, ArticleID: articleID, Name: strings.TrimSpace(articleTitle) + " Live Discussion"}
}

func (s Service) CreateRoom(ctx context.Context, articleID, name string) (domain.Room, error) {
	if strings.TrimSpace(articleID) == "" || strings.TrimSpace(name) == "" {
		return domain.Room{}, ErrInvalidInput
	}
	return s.repo.CreateRoom(ctx, domain.Room{ArticleID: articleID, Name: strings.TrimSpace(name)})
}

func (s Service) GetRoom(ctx context.Context, id string) (domain.Room, error) {
	return s.repo.GetRoom(ctx, id)
}

func (s Service) GetRoomByArticle(ctx context.Context, articleID string) (domain.Room, error) {
	return s.repo.GetRoomByArticle(ctx, articleID)
}

func (s Service) ListRooms(ctx context.Context, limit int, cursor time.Time) ([]domain.Room, error) {
	return s.repo.ListRooms(ctx, limit, cursor)
}

func (s Service) LockRoom(ctx context.Context, id string, locked bool) error {
	return s.repo.SetRoomLocked(ctx, id, locked)
}

func (s Service) ArchiveRoom(ctx context.Context, id string) error {
	return s.repo.ArchiveRoom(ctx, id)
}

func (s Service) CreateMessage(ctx context.Context, input domain.Message) (domain.Message, error) {
	if strings.TrimSpace(input.RoomID) == "" || strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.Body) == "" {
		return domain.Message{}, ErrInvalidInput
	}
	room, err := s.repo.GetRoom(ctx, input.RoomID)
	if err != nil {
		return domain.Message{}, err
	}
	if room.Locked {
		return domain.Message{}, ErrRoomLocked
	}
	if room.ArchivedAt != nil {
		return domain.Message{}, ErrRoomArchived
	}
	input.Body = strings.TrimSpace(input.Body)
	if input.MessageType == "" {
		input.MessageType = "text"
	}
	return s.repo.CreateMessage(ctx, input)
}

func (s Service) ListMessages(ctx context.Context, roomID string, limit int, cursor time.Time) ([]domain.Message, error) {
	return s.repo.ListMessages(ctx, roomID, limit, cursor)
}

func (s Service) EditMessage(ctx context.Context, messageID, userID, body string) (domain.Message, error) {
	if strings.TrimSpace(body) == "" {
		return domain.Message{}, ErrInvalidInput
	}
	return s.repo.EditMessage(ctx, messageID, userID, strings.TrimSpace(body))
}

func (s Service) DeleteMessage(ctx context.Context, messageID, userID string) error {
	return s.repo.DeleteMessage(ctx, messageID, userID)
}

func (s Service) SetPresence(ctx context.Context, presence domain.Presence) error {
	if strings.TrimSpace(presence.RoomID) == "" || strings.TrimSpace(presence.UserID) == "" {
		return ErrInvalidInput
	}
	presence.UpdatedAt = time.Now().UTC()
	return s.repo.UpsertPresence(ctx, presence)
}

func (s Service) ListPresence(ctx context.Context, roomID string) ([]domain.Presence, error) {
	return s.repo.ListPresence(ctx, roomID)
}
