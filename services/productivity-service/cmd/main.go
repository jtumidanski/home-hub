package main

import (
	"context"

	"github.com/jtumidanski/home-hub/services/productivity-service/internal/config"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "productivity-service")
	defer shutdownTracing(context.Background())

	_ = database.Connect(l, cfg.DB)

	server.New(l).
		WithAddr(":" + cfg.Port).
		Run()
}
