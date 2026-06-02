package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	AppHost         string
	AdminHost       string
	APIHost         string
	AppTarget       *url.URL
	AdminTarget     *url.URL
	APITarget       *url.URL
	AllowedOrigins  []string
	RateLimitRPS    int
	ShutdownTimeout time.Duration
	CSRFCookieName  string
	AuthCookieNames []string
	CookieSecure    bool
}

func Load() (Config, error) {
	appTarget, err := url.Parse(env("APP_TARGET", "http://frontend-user:3000"))
	if err != nil {
		return Config{}, err
	}
	adminTarget, err := url.Parse(env("ADMIN_TARGET", "http://frontend-admin:3001"))
	if err != nil {
		return Config{}, err
	}
	apiTarget, err := url.Parse(env("API_TARGET", "http://api:8080"))
	if err != nil {
		return Config{}, err
	}
	return Config{
		Port:            env("PORT", "80"),
		AppHost:         env("APP_HOST", "app.newfeed.site"),
		AdminHost:       env("ADMIN_HOST", "admin.newfeed.site"),
		APIHost:         env("API_HOST", "api.newfeed.site"),
		AppTarget:       appTarget,
		AdminTarget:     adminTarget,
		APITarget:       apiTarget,
		AllowedOrigins:  csv("CORS_ALLOWED_ORIGINS", "https://app.newfeed.site,https://admin.newfeed.site"),
		RateLimitRPS:    intEnv("RATE_LIMIT_RPS", 20),
		ShutdownTimeout: durationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
		CSRFCookieName:  env("CSRF_COOKIE_NAME", "csrf_token"),
		AuthCookieNames: csv("AUTH_COOKIE_NAMES", "newfeed_access_token,newfeed_refresh_token"),
		CookieSecure:    env("COOKIE_SECURE", "true") == "true",
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func csv(key, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func intEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
