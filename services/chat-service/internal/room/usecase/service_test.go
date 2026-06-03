package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/newfeed/community-news/services/chat-service/internal/room/domain"
)

func TestCreateMessageRejectsLockedRoom(t *testing.T) {
	service := NewService(fakeRepo{room: domain.Room{ID: "room-1", Locked: true}})

	_, err := service.CreateMessage(context.Background(), domain.Message{
		RoomID: "room-1",
		UserID: "user-1",
		Body:   "hello",
	})

	if err != ErrRoomLocked {
		t.Fatalf("expected ErrRoomLocked, got %v", err)
	}
}

func TestCreateMessageTrimsBodyAndDefaultsType(t *testing.T) {
	service := NewService(fakeRepo{room: domain.Room{ID: "room-1"}})

	message, err := service.CreateMessage(context.Background(), domain.Message{
		RoomID: "room-1",
		UserID: "user-1",
		Body:   " hello ",
	})

	if err != nil {
		t.Fatalf("create message: %v", err)
	}
	if message.Body != "hello" {
		t.Fatalf("expected trimmed body, got %q", message.Body)
	}
	if message.MessageType != "text" {
		t.Fatalf("expected default text type, got %q", message.MessageType)
	}
}

type fakeRepo struct {
	room domain.Room
}

func (f fakeRepo) CreateRoom(ctx context.Context, room domain.Room) (domain.Room, error) {
	return room, nil
}

func (f fakeRepo) GetRoom(ctx context.Context, id string) (domain.Room, error) {
	return f.room, nil
}

func (f fakeRepo) GetRoomByArticle(ctx context.Context, articleID string) (domain.Room, error) {
	return f.room, nil
}

func (f fakeRepo) ListRooms(ctx context.Context, limit int, cursor time.Time) ([]domain.Room, error) {
	return []domain.Room{f.room}, nil
}

func (f fakeRepo) SetRoomLocked(ctx context.Context, id string, locked bool) error { return nil }

func (f fakeRepo) ArchiveRoom(ctx context.Context, id string) error { return nil }

func (f fakeRepo) CreateMessage(ctx context.Context, message domain.Message) (domain.Message, error) {
	return message, nil
}

func (f fakeRepo) ListMessages(ctx context.Context, roomID string, limit int, cursor time.Time) ([]domain.Message, error) {
	return nil, nil
}

func (f fakeRepo) EditMessage(ctx context.Context, messageID, userID, body string) (domain.Message, error) {
	return domain.Message{}, nil
}

func (f fakeRepo) DeleteMessage(ctx context.Context, messageID, userID string) error { return nil }

func (f fakeRepo) UpsertPresence(ctx context.Context, presence domain.Presence) error { return nil }

func (f fakeRepo) ListPresence(ctx context.Context, roomID string) ([]domain.Presence, error) {
	return nil, nil
}
