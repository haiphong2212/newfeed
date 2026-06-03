package http

import (
	"time"

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
	app.Get("/v1/users/:user_id/articles", h.userArticles)
	app.Post("/v1/articles/:article_id/comments", h.createComment)
	app.Get("/v1/articles/:article_id/comments", h.listComments)
	app.Post("/v1/articles/:article_id/share", h.shareArticle)
	app.Get("/v1/users/:user_id/shares", h.userShares)
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

func (h *Handler) userArticles(c *fiber.Ctx) error {
	articles, err := h.articles.ListPublishedByAuthor(c.UserContext(), c.Params("user_id"), c.QueryInt("limit", 20), parseCursor(c.Query("cursor")))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"articles": articles, "next_cursor": articleCursor(articles)})
}

func (h *Handler) createComment(c *fiber.Ctx) error {
	var comment domain.Comment
	if err := c.BodyParser(&comment); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	comment.ArticleID = c.Params("article_id")
	created, err := h.articles.CreateComment(c.UserContext(), comment)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handler) listComments(c *fiber.Ctx) error {
	comments, err := h.articles.ListComments(c.UserContext(), c.Params("article_id"), c.QueryInt("limit", 50), parseCursor(c.Query("cursor")))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"comments": comments, "next_cursor": commentCursor(comments)})
}

func (h *Handler) shareArticle(c *fiber.Ctx) error {
	var share domain.Share
	if err := c.BodyParser(&share); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	share.ArticleID = c.Params("article_id")
	created, err := h.articles.ShareArticle(c.UserContext(), share)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handler) userShares(c *fiber.Ctx) error {
	shares, err := h.articles.ListSharesByUser(c.UserContext(), c.Params("user_id"), c.QueryInt("limit", 20), parseCursor(c.Query("cursor")))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"shares": shares, "next_cursor": shareCursor(shares)})
}

func parseCursor(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	cursor, _ := time.Parse(time.RFC3339Nano, value)
	return cursor
}

func articleCursor(articles []domain.Article) string {
	if len(articles) == 0 {
		return ""
	}
	return articles[len(articles)-1].CreatedAt.Format(time.RFC3339Nano)
}

func commentCursor(comments []domain.Comment) string {
	if len(comments) == 0 {
		return ""
	}
	return comments[len(comments)-1].CreatedAt.Format(time.RFC3339Nano)
}

func shareCursor(shares []domain.Share) string {
	if len(shares) == 0 {
		return ""
	}
	return shares[len(shares)-1].CreatedAt.Format(time.RFC3339Nano)
}
