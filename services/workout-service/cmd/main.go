package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/config"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/region"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/summary"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/theme"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/today"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/weekview"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "workout-service")
	defer shutdownTracing(context.Background())

	// Migration order matters: parents (themes, regions, exercises, weeks)
	// must exist before children (planned_items, performances) so the
	// post-AutoMigrate FK ALTERs in the child packages succeed.
	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			theme.Migration,
			region.Migration,
			exercise.Migration,
			week.Migration,
			planneditem.Migration,
			performance.Migration,
		),
	)

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			// Order matters: more specific routes before generic
			// /workouts/weeks/{weekStart}/... routes.
			today.InitializeRoutes(db, cfg.AccountBaseURL)(l, si, api)
			summary.InitializeRoutes(db)(l, si, api)
			performance.InitializeRoutes(db)(l, si, api)
			weekview.InitializeRoutes(db)(l, si, api)
			planneditem.InitializeRoutes(db)(l, si, api)
			week.InitializeRoutes(db)(l, si, api)
			exercise.InitializeRoutes(db)(l, si, api)
			region.InitializeRoutes(db)(l, si, api)
			theme.InitializeRoutes(db)(l, si, api)
		}).
		Run()
}
