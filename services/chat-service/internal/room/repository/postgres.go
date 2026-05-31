package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) SaveMessage(ctx context.Context, roomID, userID, body string) (Message, error) {
	message := Message{ID: newID(), RoomID: roomID, UserID: userID, Body: body, CreatedAt: time.Now().UTC()}
	_, err := r.db.Exec(ctx, `
		INSERT INTO chat_messages (id, room_id, user_id, body, created_at)
		VALUES ($1,$2,$3,$4,$5)
	`, message.ID, message.RoomID, message.UserID, message.Body, message.CreatedAt)
	return message, err
}

func newID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return hex.EncodeToString(buf[0:4]) + "-" + hex.EncodeToString(buf[4:6]) + "-" + hex.EncodeToString(buf[6:8]) + "-" + hex.EncodeToString(buf[8:10]) + "-" + hex.EncodeToString(buf[10:])
}
