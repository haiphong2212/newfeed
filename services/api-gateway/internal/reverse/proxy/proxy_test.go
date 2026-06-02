package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/logger"
)

func TestGatewayRoutesByHostAndForwardsHeaders(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Fatalf("expected forwarded path /users, got %s", r.URL.Path)
		}
		if r.Header.Get("X-Forwarded-Host") != "api.newfeed.site" {
			t.Fatalf("missing forwarded host: %q", r.Header.Get("X-Forwarded-Host"))
		}
		if r.Header.Get("X-Real-IP") == "" {
			t.Fatal("missing X-Real-IP")
		}
		w.Header().Set("X-Upstream", "api")
		w.WriteHeader(http.StatusCreated)
	}))
	defer upstream.Close()

	target, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	gateway := NewGateway(map[string]*url.URL{"api.newfeed.site": target}, logger.New())
	req := httptest.NewRequest(http.MethodGet, "http://api.newfeed.site/users", nil)
	req.Host = "api.newfeed.site"
	rec := httptest.NewRecorder()

	gateway.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get("X-Upstream") != "api" {
		t.Fatal("response was not proxied")
	}
}

func TestGatewayRejectsUnknownHost(t *testing.T) {
	gateway := NewGateway(map[string]*url.URL{}, logger.New())
	req := httptest.NewRequest(http.MethodGet, "http://unknown.local", nil)
	rec := httptest.NewRecorder()

	gateway.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}
