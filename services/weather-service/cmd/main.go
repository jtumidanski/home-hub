package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/config"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/forecast"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/geocoding"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/refresh"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "weather-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			forecast.Migration,
		),
	)

	client := openmeteo.NewClient()
	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go refresh.StartRefreshLoop(ctx, db, client, cfg.RefreshInterval, l)

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			forecast.InitializeRoutes(db, client)(l, si, api)
			geocoding.InitializeRoutes(client)(l, si, api)
		}).
		Run()
}
