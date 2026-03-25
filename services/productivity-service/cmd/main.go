package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/config"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder/dismissal"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder/snooze"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/summary"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task/restoration"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "productivity-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			task.Migration,
			restoration.Migration,
			reminder.Migration,
			snooze.Migration,
			dismissal.Migration,
		),
	)

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			task.InitializeRoutes(db)(l, si, api)
			restoration.InitializeRoutes(db)(l, si, api)
			reminder.InitializeRoutes(db)(l, si, api)
			snooze.InitializeRoutes(db)(l, si, api)
			dismissal.InitializeRoutes(db)(l, si, api)
			summary.InitializeRoutes(db)(l, si, api)
		}).
		Run()
}
