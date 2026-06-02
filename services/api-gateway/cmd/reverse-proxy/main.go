package main

import (
	"log"

	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/config"
	gatewaylog "github.com/newfeed/community-news/services/api-gateway/internal/reverse/logger"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/router"
	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := gatewaylog.New()
	handler := router.New(cfg, logger)
	if err := server.Run(":"+cfg.Port, handler, cfg.ShutdownTimeout, logger); err != nil {
		log.Fatal(err)
	}
}
