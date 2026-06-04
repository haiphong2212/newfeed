package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/newfeed/community-news/services/chat-service/internal/room/domain"
	"github.com/redis/go-redis/v9"
)

type Repository interface {
	SaveMessage(ctx context.Context, roomID, userID, body string) (domain.Message, error)
}

type Hub struct {
	repo  Repository
	redis *redis.Client
	mu    sync.RWMutex
	rooms map[string]map[*fiberws.Conn]struct{}
}

func NewHub(repo Repository, redis *redis.Client) *Hub {
	return &Hub{repo: repo, redis: redis, rooms: map[string]map[*fiberws.Conn]struct{}{}}
}

func (h *Hub) RegisterRoutes(app *fiber.App) {
	app.Get("/ws/rooms/:room_id", fiberws.New(h.handle))
}

func (h *Hub) handle(conn *fiberws.Conn) {
	roomID := conn.Params("room_id")
	userID := conn.Query("user_id")
	h.join(roomID, userID, conn)
	defer h.leave(roomID, userID, conn)

	for {
		var input struct {
			Type string `json:"type"`
			Body string `json:"body"`
		}
		if err := conn.ReadJSON(&input); err != nil {
			return
		}
		if input.Type == "typing" {
			h.setPresence(roomID, userID, true, true)
			h.broadcast(roomID, map[string]any{"type": "typing", "room_id": roomID, "user_id": userID})
			continue
		}
		message, err := h.repo.SaveMessage(context.Background(), roomID, userID, input.Body)
		if err != nil {
			_ = conn.WriteJSON(map[string]any{"type": "error", "error": err.Error()})
			continue
		}
		h.broadcast(roomID, map[string]any{"type": "message", "message": message})
	}
}

func (h *Hub) join(roomID, userID string, conn *fiberws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = map[*fiberws.Conn]struct{}{}
	}
	h.rooms[roomID][conn] = struct{}{}
	h.setPresence(roomID, userID, true, false)
}

func (h *Hub) leave(roomID, userID string, conn *fiberws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms[roomID], conn)
	_ = conn.Close()
	h.setPresence(roomID, userID, false, false)
}

func (h *Hub) broadcast(roomID string, payload any) {
	data, _ := json.Marshal(payload)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.rooms[roomID] {
		_ = conn.WriteMessage(fiberws.TextMessage, data)
	}
}

func (h *Hub) setPresence(roomID, userID string, online, typing bool) {
	if userID == "" {
		return
	}
	_ = h.redis.HSet(context.Background(), "room_presence:"+roomID+":"+userID, map[string]any{
		"online":     online,
		"typing":     typing,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}).Err()
}
