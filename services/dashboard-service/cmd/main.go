package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/config"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/dashboard"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/retention"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "dashboard-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(dashboard.Migration),
	)

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			if _, err := retention.Setup(ctx, l, db, router, cfg.AccountServiceURL, cfg.InternalToken, cfg.RetentionInterval); err != nil {
				l.WithError(err).Fatal("retention setup failed")
			}

			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			dashboard.InitializeRoutes(db)(l, si, api)
		}).
		Run()
}
