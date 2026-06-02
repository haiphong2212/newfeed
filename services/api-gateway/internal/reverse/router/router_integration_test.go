package router

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/config"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/logger"
)

func TestRouterIntegrationProxiesAPIHost(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer api.Close()

	apiURL, err := url.Parse(api.URL)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		AppHost:         "app.newfeed.site",
		AdminHost:       "admin.newfeed.site",
		APIHost:         "api.newfeed.site",
		AppTarget:       apiURL,
		AdminTarget:     apiURL,
		APITarget:       apiURL,
		AllowedOrigins:  []string{"https://app.newfeed.site"},
		RateLimitRPS:    100,
		CSRFCookieName:  "csrf_token",
		AuthCookieNames: []string{"newfeed_access_token"},
	}
	handler := New(cfg, logger.New())
	req := httptest.NewRequest(http.MethodGet, "http://api.newfeed.site/users", nil)
	req.Host = "api.newfeed.site"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 from upstream, got %d", rec.Code)
	}
}

func TestRouterHealth(t *testing.T) {
	target, _ := url.Parse("http://127.0.0.1:1")
	cfg := config.Config{
		AppHost:         "app.newfeed.site",
		AdminHost:       "admin.newfeed.site",
		APIHost:         "api.newfeed.site",
		AppTarget:       target,
		AdminTarget:     target,
		APITarget:       target,
		AllowedOrigins:  []string{"*"},
		RateLimitRPS:    100,
		CSRFCookieName:  "csrf_token",
		AuthCookieNames: []string{"newfeed_access_token"},
	}
	handler := New(cfg, logger.New())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
