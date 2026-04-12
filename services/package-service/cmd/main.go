package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/package-service/internal/carrier"
	"github.com/jtumidanski/home-hub/services/package-service/internal/config"
	"github.com/jtumidanski/home-hub/services/package-service/internal/poller"
	"github.com/jtumidanski/home-hub/services/package-service/internal/retention"
	"github.com/jtumidanski/home-hub/services/package-service/internal/tracking"
	"github.com/jtumidanski/home-hub/services/package-service/internal/trackingevent"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "package-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			tracking.Migration,
			trackingevent.Migration,
		),
	)

	httpClient := carrier.NewHTTPClient()
	tokenMgr := carrier.NewOAuthTokenManager(httpClient, l)
	budget := carrier.NewRateBudget(map[string]int{
		"usps":  1000,
		"ups":   250,
		"fedex": 500,
	})

	carriers := carrier.NewRegistry()
	if cfg.USPSClientID != "" {
		carriers.Register(carrier.NewUSPSClient(cfg.USPSClientID, cfg.USPSClientSecret, tokenMgr, budget, httpClient, l))
	}
	if cfg.UPSClientID != "" {
		carriers.Register(carrier.NewUPSClient(cfg.UPSClientID, cfg.UPSClientSecret, tokenMgr, budget, httpClient, l))
	}
	if cfg.FedExAPIKey != "" {
		carriers.Register(carrier.NewFedExClient(cfg.FedExAPIKey, cfg.FedExSecretKey, cfg.FedExSandbox, tokenMgr, budget, httpClient, l))
	}

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	pollEngine := poller.NewEngine(db, carriers, cfg.MaxActivePerHousehold, l)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pollEngine.StartPollLoop(ctx, cfg.PollInterval, cfg.PollIntervalUrgent)
	go pollEngine.StartCleanupLoop(ctx, poller.CleanupConfig{
		StaleAfterDays: cfg.StaleAfterDays,
	})

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			if _, err := retention.Setup(ctx, l, db, router, cfg.AccountServiceURL, cfg.InternalToken, cfg.RetentionInterval); err != nil {
				l.WithError(err).Fatal("retention setup failed")
			}

			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			tracking.InitializeRoutes(db, cfg.MaxActivePerHousehold, carriers)(l, si, api)
		}).
		Run()
}
