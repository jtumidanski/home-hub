package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/config"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/month"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/retention"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/today"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "tracker-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			trackingitem.Migration,
			schedule.Migration,
			entry.Migration,
		),
	)

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			if _, err := retention.Setup(ctx, l, db, router, cfg.AccountBaseURL, cfg.InternalToken, cfg.RetentionInterval); err != nil {
				l.WithError(err).Fatal("retention setup failed")
			}

			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			// Order matters: more specific routes first
			today.InitializeRoutes(db, cfg.AccountBaseURL)(l, si, api)
			month.InitializeRoutes(db)(l, si, api)
			entry.InitializeRoutes(db)(l, si, api)
			trackingitem.InitializeRoutes(db)(l, si, api)
		}).
		Run()
}
