package http

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	gatewaygrpc "github.com/newfeed/community-news/services/api-gateway/internal/gateway/grpc"
)

func TestRegisterRoutesAcceptsAPIPrefix(t *testing.T) {
	app := fiber.New()
	NewHandler(gatewaygrpc.NewHealthClient(map[string]string{}), nil, HTTPTargets{}).RegisterRoutes(app)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/status", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}
