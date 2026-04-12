package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/categoryclient"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/config"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/ingredient"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/plan"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planner"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/retention"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "recipe-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			recipe.Migration,
			ingredient.Migration,
			normalization.Migration,
			planner.Migration,
			audit.Migration,
			plan.Migration,
			planitem.Migration,
		),
	)

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()
	catClient := categoryclient.New(cfg.CategoryServiceURL)

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

			recipe.InitializeRoutes(db)(l, si, api)
			ingredient.InitializeRoutes(db, catClient)(l, si, api)
			normalization.InitializeRoutes(db)(l, si, api)
			plan.InitializeRoutes(db, catClient)(l, si, api)
		}).
		Run()
}
