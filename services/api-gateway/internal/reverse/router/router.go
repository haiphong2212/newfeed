package router

import (
	"net/http"
	"net/url"

	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/config"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/logger"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/middleware"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/proxy"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/response"
)

func New(cfg config.Config, log *logger.Logger) http.Handler {
	gateway := proxy.NewGateway(map[string]*url.URL{
		cfg.AppHost:   cfg.AppTarget,
		cfg.AdminHost: cfg.AdminTarget,
		cfg.APIHost:   cfg.APITarget,
	}, log)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("/csrf-token", middleware.TokenHandler(cfg.CSRFCookieName, cfg.CookieSecure))
	mux.Handle("/api/", proxy.NewPrefixProxy(cfg.APITarget, "/api", log))
	mux.Handle("/ws/", proxy.NewPrefixProxy(cfg.ChatTarget, "", log))
	mux.Handle("/", gateway)
	return middleware.Chain(
		mux,
		middleware.RequestID(),
		middleware.AccessLog(log),
		middleware.Recovery(log),
		middleware.SecurityHeaders(),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.RateLimit(cfg.RateLimitRPS),
		middleware.CSRF(middleware.CSRFOptions{
			CookieName:      cfg.CSRFCookieName,
			AuthCookieNames: cfg.AuthCookieNames,
			CookieSecure:    cfg.CookieSecure,
		}),
	)
}
