package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/account-service/internal/appcontext"
	"github.com/jtumidanski/home-hub/services/account-service/internal/config"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/invitation"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/services/account-service/internal/tenant"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "account-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			tenant.Migration,
			household.Migration,
			membership.Migration,
			preference.Migration,
			invitation.Migration,
		),
	)

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			tenant.InitializeRoutes(db)(l, si, api)
			household.InitializeRoutes(db)(l, si, api)
			membership.InitializeRoutes(db)(l, si, api)
			preference.InitializeRoutes(db)(l, si, api)
			invitation.InitializeRoutes(db)(l, si, api)
			appcontext.InitializeRoutes(db)(l, si, api)
		}).
		Run()
}
