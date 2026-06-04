package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newfeed/community-news/services/chat-service/internal/room/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateRoom(ctx context.Context, room domain.Room) (domain.Room, error) {
	if room.ID == "" {
		room.ID = newID()
	}
	if room.CreatedAt.IsZero() {
		room.CreatedAt = time.Now().UTC()
	}
	err := r.db.QueryRow(ctx, `
		INSERT INTO discussion_rooms (id, article_id, name, locked, archived_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (article_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, article_id, name, locked, archived_at, created_at
	`, room.ID, room.ArticleID, room.Name, room.Locked, room.ArchivedAt, room.CreatedAt).Scan(
		&room.ID, &room.ArticleID, &room.Name, &room.Locked, &room.ArchivedAt, &room.CreatedAt,
	)
	return room, err
}

func (r *PostgresRepository) GetRoom(ctx context.Context, id string) (domain.Room, error) {
	var room domain.Room
	err := r.db.QueryRow(ctx, `
		SELECT id, article_id, name, locked, archived_at, created_at
		FROM discussion_rooms
		WHERE id = $1
	`, id).Scan(&room.ID, &room.ArticleID, &room.Name, &room.Locked, &room.ArchivedAt, &room.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Room{}, ErrNotFound
	}
	return room, err
}

func (r *PostgresRepository) GetRoomByArticle(ctx context.Context, articleID string) (domain.Room, error) {
	var room domain.Room
	err := r.db.QueryRow(ctx, `
		SELECT id, article_id, name, locked, archived_at, created_at
		FROM discussion_rooms
		WHERE article_id = $1
	`, articleID).Scan(&room.ID, &room.ArticleID, &room.Name, &room.Locked, &room.ArchivedAt, &room.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Room{}, ErrNotFound
	}
	return room, err
}

func (r *PostgresRepository) ListRooms(ctx context.Context, limit int, cursor time.Time) ([]domain.Room, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if cursor.IsZero() {
		cursor = time.Now().UTC().Add(100 * 365 * 24 * time.Hour)
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, article_id, name, locked, archived_at, created_at
		FROM discussion_rooms
		WHERE created_at < $1
		ORDER BY created_at DESC
		LIMIT $2
	`, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms := make([]domain.Room, 0)
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.ArticleID, &room.Name, &room.Locked, &room.ArchivedAt, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (r *PostgresRepository) SetRoomLocked(ctx context.Context, id string, locked bool) error {
	tag, err := r.db.Exec(ctx, `UPDATE discussion_rooms SET locked = $2 WHERE id = $1`, id, locked)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ArchiveRoom(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `UPDATE discussion_rooms SET archived_at = now() WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) SaveMessage(ctx context.Context, roomID, userID, body string) (domain.Message, error) {
	return r.CreateMessage(ctx, domain.Message{
		RoomID:      roomID,
		UserID:      userID,
		MessageType: "text",
		Body:        body,
		Metadata:    map[string]any{},
	})
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, message domain.Message) (domain.Message, error) {
	if message.ID == "" {
		message.ID = newID()
	}
	if message.MessageType == "" {
		message.MessageType = "text"
	}
	if message.Metadata == nil {
		message.Metadata = map[string]any{}
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}
	metadata, _ := json.Marshal(message.Metadata)
	err := r.db.QueryRow(ctx, `
		INSERT INTO chat_messages (id, room_id, user_id, message_type, body, metadata, edited_at, deleted_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, room_id, user_id, message_type, body, metadata, edited_at, deleted_at, created_at
	`, message.ID, message.RoomID, message.UserID, message.MessageType, message.Body, metadata, message.EditedAt, message.DeletedAt, message.CreatedAt).Scan(
		&message.ID, &message.RoomID, &message.UserID, &message.MessageType, &message.Body, &metadata, &message.EditedAt, &message.DeletedAt, &message.CreatedAt,
	)
	if err != nil {
		return domain.Message{}, err
	}
	_ = json.Unmarshal(metadata, &message.Metadata)
	return message, nil
}

func (r *PostgresRepository) ListMessages(ctx context.Context, roomID string, limit int, cursor time.Time) ([]domain.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if cursor.IsZero() {
		cursor = time.Now().UTC().Add(100 * 365 * 24 * time.Hour)
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, room_id, user_id, message_type, body, metadata, edited_at, deleted_at, created_at
		FROM chat_messages
		WHERE room_id = $1 AND created_at < $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3
	`, roomID, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]domain.Message, 0)
	for rows.Next() {
		var message domain.Message
		var metadata []byte
		if err := rows.Scan(&message.ID, &message.RoomID, &message.UserID, &message.MessageType, &message.Body, &metadata, &message.EditedAt, &message.DeletedAt, &message.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metadata, &message.Metadata)
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (r *PostgresRepository) EditMessage(ctx context.Context, messageID, userID, body string) (domain.Message, error) {
	var message domain.Message
	var metadata []byte
	err := r.db.QueryRow(ctx, `
		UPDATE chat_messages
		SET body = $3, edited_at = now()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, room_id, user_id, message_type, body, metadata, edited_at, deleted_at, created_at
	`, messageID, userID, body).Scan(&message.ID, &message.RoomID, &message.UserID, &message.MessageType, &message.Body, &metadata, &message.EditedAt, &message.DeletedAt, &message.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Message{}, ErrNotFound
	}
	_ = json.Unmarshal(metadata, &message.Metadata)
	return message, err
}

func (r *PostgresRepository) DeleteMessage(ctx context.Context, messageID, userID string) error {
	tag, err := r.db.Exec(ctx, `UPDATE chat_messages SET deleted_at = now() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, messageID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) UpsertPresence(ctx context.Context, presence domain.Presence) error {
	if presence.UpdatedAt.IsZero() {
		presence.UpdatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO room_presence (room_id, user_id, online, typing, updated_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (room_id, user_id) DO UPDATE SET
			online = EXCLUDED.online,
			typing = EXCLUDED.typing,
			updated_at = EXCLUDED.updated_at
	`, presence.RoomID, presence.UserID, presence.Online, presence.Typing, presence.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListPresence(ctx context.Context, roomID string) ([]domain.Presence, error) {
	rows, err := r.db.Query(ctx, `
		SELECT room_id, user_id, online, typing, updated_at
		FROM room_presence
		WHERE room_id = $1 AND online = TRUE
		ORDER BY updated_at DESC
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	presence := make([]domain.Presence, 0)
	for rows.Next() {
		var item domain.Presence
		if err := rows.Scan(&item.RoomID, &item.UserID, &item.Online, &item.Typing, &item.UpdatedAt); err != nil {
			return nil, err
		}
		presence = append(presence, item)
	}
	return presence, rows.Err()
}

var ErrNotFound = errors.New("resource not found")

func newID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return hex.EncodeToString(buf[0:4]) + "-" + hex.EncodeToString(buf[4:6]) + "-" + hex.EncodeToString(buf[6:8]) + "-" + hex.EncodeToString(buf[8:10]) + "-" + hex.EncodeToString(buf[10:])
}
