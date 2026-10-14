package main

import (
	"context"
	"time"

	"github.com/gorilla/mux"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/categoryclient"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/config"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/ingredient"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/plan"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planner"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/retention"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "recipe-service")
	defer func() { _ = shutdownTracing(context.Background()) }()

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			recipe.Migration,
			ingredient.Migration,
			normalization.Migration,
			planner.Migration,
			audit.Migration,
			plan.Migration,
			planitem.Migration,
			recipeimage.Migration,
		),
	)

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()
	catClient := categoryclient.New(cfg.CategoryServiceURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := objectstore.New(l, cfg.ObjectStore)
	go func() {
		pctx, pcancel := context.WithTimeout(ctx, 5*time.Second)
		defer pcancel()
		if err := store.Ping(pctx); err != nil {
			recipeimage.RecordStorageError("ping")
			l.WithError(err).Error("object storage unreachable at startup; image endpoints will return 503 until it recovers")
		}
	}()
	imageReaper := recipeimage.NewObjectReaper(l, db, store)
	go sr.Loop(ctx, l, 5*time.Minute, imageReaper.Run)

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
			planitem.InitializeRoutes(db, plan.NewPlanProvider(db), plan.NewRecipeValidator(db))(l, si, api)
			recipeimage.InitializeRoutes(db, store, recipe.ImageVisibility(db), imageReaper)(l, si, api)
		}).
		Run()
}
