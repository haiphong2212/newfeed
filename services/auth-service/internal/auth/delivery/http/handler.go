package http

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/usecase"
)

type Handler struct {
	auth    *usecase.Service
	cookies CookieOptions
}

type CookieOptions struct {
	Enabled             bool
	Secure              bool
	Domain              string
	SameSite            string
	AccessCookieName    string
	RefreshCookieName   string
	RefreshCookieMaxAge time.Duration
}

func NewHandler(auth *usecase.Service, cookies CookieOptions) *Handler {
	if cookies.AccessCookieName == "" {
		cookies.AccessCookieName = "newfeed_access_token"
	}
	if cookies.RefreshCookieName == "" {
		cookies.RefreshCookieName = "newfeed_refresh_token"
	}
	if cookies.SameSite == "" {
		cookies.SameSite = "Lax"
	}
	return &Handler{auth: auth, cookies: cookies}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/v1/auth/register", h.register)
	app.Post("/v1/auth/login", h.login)
	app.Post("/v1/auth/refresh", h.refresh)
	app.Post("/v1/auth/logout", h.logout)
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
	h.setAuthCookies(c, tokens)
	return c.JSON(tokens)
}

func (h *Handler) refresh(c *fiber.Ctx) error {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	if input.RefreshToken == "" {
		input.RefreshToken = c.Cookies(h.cookies.RefreshCookieName)
	}
	tokens, err := h.auth.Refresh(c.UserContext(), input.RefreshToken)
	if err != nil {
		return authError(err)
	}
	h.setAuthCookies(c, tokens)
	return c.JSON(tokens)
}

func (h *Handler) validate(c *fiber.Ctx) error {
	raw := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	if raw == "" {
		raw = c.Cookies(h.cookies.AccessCookieName)
	}
	claims, err := h.auth.ValidateAccessToken(c.UserContext(), raw)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}
	return c.JSON(claims)
}

func (h *Handler) logout(c *fiber.Ctx) error {
	h.clearCookie(c, h.cookies.AccessCookieName)
	h.clearCookie(c, h.cookies.RefreshCookieName)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) setAuthCookies(c *fiber.Ctx, tokens *domain.TokenPair) {
	if !h.cookies.Enabled {
		return
	}
	h.setCookie(c, h.cookies.AccessCookieName, tokens.AccessToken, time.Duration(tokens.ExpiresIn)*time.Second)
	h.setCookie(c, h.cookies.RefreshCookieName, tokens.RefreshToken, h.cookies.RefreshCookieMaxAge)
}

func (h *Handler) setCookie(c *fiber.Ctx, name, value string, maxAge time.Duration) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   h.cookies.Domain,
		MaxAge:   int(maxAge.Seconds()),
		Secure:   h.cookies.Secure,
		HTTPOnly: true,
		SameSite: h.cookies.SameSite,
	})
}

func (h *Handler) clearCookie(c *fiber.Ctx, name string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   h.cookies.Domain,
		MaxAge:   -1,
		Secure:   h.cookies.Secure,
		HTTPOnly: true,
		SameSite: h.cookies.SameSite,
	})
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
