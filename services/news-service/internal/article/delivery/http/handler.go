package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/newfeed/community-news/services/news-service/internal/article/domain"
	"github.com/newfeed/community-news/services/news-service/internal/article/usecase"
)

type Handler struct {
	articles *usecase.Service
}

func NewHandler(articles *usecase.Service) *Handler {
	return &Handler{articles: articles}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/v1/articles/publish", h.publish)
}

func (h *Handler) publish(c *fiber.Ctx) error {
	var article domain.Article
	if err := c.BodyParser(&article); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	if err := h.articles.Publish(c.UserContext(), article); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "published", "discussion_room_name": article.DiscussionRoomName()})
}
