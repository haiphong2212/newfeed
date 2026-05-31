package http

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/usecase"
)

type Handler struct {
	auth *usecase.Service
}

func NewHandler(auth *usecase.Service) *Handler {
	return &Handler{auth: auth}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/v1/auth/register", h.register)
	app.Post("/v1/auth/login", h.login)
	app.Post("/v1/auth/refresh", h.refresh)
	app.Get("/v1/auth/validate", h.validate)
}

func (h *Handler) register(c *fiber.Ctx) error {
	var input usecase.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	user, err := h.auth.Register(c.UserContext(), input)
	if err != nil {
		return authError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
	})
}

func (h *Handler) login(c *fiber.Ctx) error {
	var input usecase.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	tokens, err := h.auth.Login(c.UserContext(), input)
	if err != nil {
		return authError(err)
	}
	return c.JSON(tokens)
}

func (h *Handler) refresh(c *fiber.Ctx) error {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	tokens, err := h.auth.Refresh(c.UserContext(), input.RefreshToken)
	if err != nil {
		return authError(err)
	}
	return c.JSON(tokens)
}

func (h *Handler) validate(c *fiber.Ctx) error {
	raw := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	claims, err := h.auth.ValidateAccessToken(c.UserContext(), raw)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}
	return c.JSON(claims)
}

func authError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidPassword):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrRefreshNotFound):
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
}
