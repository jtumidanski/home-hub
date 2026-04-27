package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/account-service/internal/appcontext"
	"github.com/jtumidanski/home-hub/services/account-service/internal/config"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/householdpreference"
	"github.com/jtumidanski/home-hub/services/account-service/internal/invitation"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/services/account-service/internal/retention"
	"github.com/jtumidanski/home-hub/services/account-service/internal/tenant"
	"github.com/jtumidanski/home-hub/services/account-service/internal/userlifecycle"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	kprod "github.com/jtumidanski/home-hub/shared/go/kafka/producer"
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
			householdpreference.Migration,
			invitation.Migration,
			retention.Migration,
		),
	)

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()
	fan := retention.NewHTTPFanout(cfg.ServiceURLs, cfg.InternalToken, l)

	// Kafka producer for cross-service events. Produce retries + logs;
	// broker-down at startup is not fatal — the event is dropped with a warn.
	var userLifecycleProducer userlifecycle.Producer
	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" {
		userLifecycleProducer = kprod.New(kprod.Config{Brokers: cfg.KafkaBrokers}, l)
	}

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			retention.InitializeInternalRoutes(db, cfg.InternalToken)(l, router)
			userlifecycle.InitializeRoutes(db, userLifecycleProducer, userlifecycle.Config{
				Topic:         cfg.TopicUserLifecycle,
				InternalToken: cfg.InternalToken,
			})(l, router)

			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			tenant.InitializeRoutes(db)(l, si, api)
			household.InitializeRoutes(db)(l, si, api)
			membership.InitializeRoutes(db)(l, si, api)
			preference.InitializeRoutes(db)(l, si, api)
			householdpreference.InitializeRoutes(db)(l, si, api)
			invitation.InitializeRoutes(db)(l, si, api)
			appcontext.InitializeRoutes(db)(l, si, api)
			retention.InitializeRoutes(db, fan)(l, si, api)
		}).
		Run()
}
