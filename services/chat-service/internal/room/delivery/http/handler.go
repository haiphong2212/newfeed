package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/newfeed/community-news/services/chat-service/internal/room/domain"
	"github.com/newfeed/community-news/services/chat-service/internal/room/repository"
	"github.com/newfeed/community-news/services/chat-service/internal/room/usecase"
)

type Handler struct {
	rooms usecase.Service
}

func NewHandler(rooms usecase.Service) *Handler {
	return &Handler{rooms: rooms}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/v1/chat/rooms", h.createRoom)
	app.Get("/v1/chat/rooms", h.listRooms)
	app.Get("/v1/chat/rooms/:room_id", h.getRoom)
	app.Get("/v1/chat/articles/:article_id/room", h.getRoomByArticle)
	app.Patch("/v1/chat/rooms/:room_id/lock", h.lockRoom)
	app.Patch("/v1/chat/rooms/:room_id/archive", h.archiveRoom)
	app.Post("/v1/chat/rooms/:room_id/messages", h.createMessage)
	app.Get("/v1/chat/rooms/:room_id/messages", h.listMessages)
	app.Patch("/v1/chat/messages/:message_id", h.editMessage)
	app.Delete("/v1/chat/messages/:message_id", h.deleteMessage)
	app.Put("/v1/chat/rooms/:room_id/presence", h.setPresence)
	app.Get("/v1/chat/rooms/:room_id/presence", h.listPresence)
}

func (h *Handler) createRoom(c *fiber.Ctx) error {
	var input struct {
		ArticleID string `json:"article_id"`
		Name      string `json:"name"`
	}
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	room, err := h.rooms.CreateRoom(c.UserContext(), input.ArticleID, input.Name)
	if err != nil {
		return writeError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(room)
}

func (h *Handler) listRooms(c *fiber.Ctx) error {
	rooms, err := h.rooms.ListRooms(c.UserContext(), c.QueryInt("limit", 20), parseCursor(c.Query("cursor")))
	if err != nil {
		return writeError(err)
	}
	return c.JSON(fiber.Map{"rooms": rooms, "next_cursor": nextRoomCursor(rooms)})
}

func (h *Handler) getRoom(c *fiber.Ctx) error {
	room, err := h.rooms.GetRoom(c.UserContext(), c.Params("room_id"))
	if err != nil {
		return writeError(err)
	}
	return c.JSON(room)
}

func (h *Handler) getRoomByArticle(c *fiber.Ctx) error {
	room, err := h.rooms.GetRoomByArticle(c.UserContext(), c.Params("article_id"))
	if err != nil {
		return writeError(err)
	}
	return c.JSON(room)
}

func (h *Handler) lockRoom(c *fiber.Ctx) error {
	var input struct {
		Locked bool `json:"locked"`
	}
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	if err := h.rooms.LockRoom(c.UserContext(), c.Params("room_id"), input.Locked); err != nil {
		return writeError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) archiveRoom(c *fiber.Ctx) error {
	if err := h.rooms.ArchiveRoom(c.UserContext(), c.Params("room_id")); err != nil {
		return writeError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) createMessage(c *fiber.Ctx) error {
	var input domain.Message
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	input.RoomID = c.Params("room_id")
	message, err := h.rooms.CreateMessage(c.UserContext(), input)
	if err != nil {
		return writeError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(message)
}

func (h *Handler) listMessages(c *fiber.Ctx) error {
	messages, err := h.rooms.ListMessages(c.UserContext(), c.Params("room_id"), c.QueryInt("limit", 50), parseCursor(c.Query("cursor")))
	if err != nil {
		return writeError(err)
	}
	return c.JSON(fiber.Map{"messages": messages, "next_cursor": nextMessageCursor(messages)})
}

func (h *Handler) editMessage(c *fiber.Ctx) error {
	var input struct {
		UserID string `json:"user_id"`
		Body   string `json:"body"`
	}
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	message, err := h.rooms.EditMessage(c.UserContext(), c.Params("message_id"), input.UserID, input.Body)
	if err != nil {
		return writeError(err)
	}
	return c.JSON(message)
}

func (h *Handler) deleteMessage(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is required")
	}
	if err := h.rooms.DeleteMessage(c.UserContext(), c.Params("message_id"), userID); err != nil {
		return writeError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) setPresence(c *fiber.Ctx) error {
	var input domain.Presence
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	input.RoomID = c.Params("room_id")
	if err := h.rooms.SetPresence(c.UserContext(), input); err != nil {
		return writeError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) listPresence(c *fiber.Ctx) error {
	presence, err := h.rooms.ListPresence(c.UserContext(), c.Params("room_id"))
	if err != nil {
		return writeError(err)
	}
	return c.JSON(fiber.Map{"presence": presence})
}

func parseCursor(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	cursor, _ := time.Parse(time.RFC3339Nano, value)
	return cursor
}

func nextRoomCursor(rooms []domain.Room) string {
	if len(rooms) == 0 {
		return ""
	}
	return rooms[len(rooms)-1].CreatedAt.Format(time.RFC3339Nano)
}

func nextMessageCursor(messages []domain.Message) string {
	if len(messages) == 0 {
		return ""
	}
	return messages[len(messages)-1].CreatedAt.Format(time.RFC3339Nano)
}

func writeError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, usecase.ErrRoomLocked), errors.Is(err, usecase.ErrRoomArchived):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	default:
		return err
	}
}
