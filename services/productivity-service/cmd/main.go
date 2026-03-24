package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/config"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminderdismissal"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/remindersnooze"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/summary"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/taskrestoration"
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
			taskrestoration.Migration,
			reminder.Migration,
			remindersnooze.Migration,
			reminderdismissal.Migration,
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
			taskrestoration.InitializeRoutes(db)(l, si, api)
			reminder.InitializeRoutes(db)(l, si, api)
			remindersnooze.InitializeRoutes(db)(l, si, api)
			reminderdismissal.InitializeRoutes(db)(l, si, api)
			summary.InitializeRoutes(db)(l, si, api)
		}).
		Run()
}
