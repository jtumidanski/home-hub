package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/config"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/forecast"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/geocoding"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/locationofinterest"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/refresh"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "weather-service")
	defer shutdownTracing(context.Background())

	// Migration order matters: locations_of_interest must exist before
	// forecast.Migration adds the FK on weather_caches.location_id.
	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			locationofinterest.Migration,
			forecast.Migration,
		),
	)

	client := openmeteo.NewClient()
	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go refresh.StartRefreshLoop(ctx, db, client, cfg.RefreshInterval, l)

	makeWarmer := func(reqL logrus.FieldLogger, r *http.Request) locationofinterest.CacheWarmer {
		return forecast.NewProcessor(reqL, r.Context(), db, client, cfg.CacheTTL, nil)
	}

	makeResolver := func(reqL logrus.FieldLogger, r *http.Request) forecast.LocationResolver {
		return locationResolverAdapter{
			proc: locationofinterest.NewProcessor(reqL, r.Context(), db, nil),
		}
	}

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			forecast.InitializeRoutes(db, client, cfg.CacheTTL, makeResolver)(l, si, api)
			geocoding.InitializeRoutes(client)(l, si, api)
			locationofinterest.InitializeRoutes(db, makeWarmer)(l, si, api)
		}).
		Run()
}

// locationResolverAdapter bridges locationofinterest.Processor to the
// forecast.LocationResolver interface, avoiding an import cycle.
type locationResolverAdapter struct {
	proc *locationofinterest.Processor
}

func (a locationResolverAdapter) ResolveLocation(householdID, locationID uuid.UUID) (float64, float64, error) {
	m, err := a.proc.Get(householdID, locationID)
	if err != nil {
		if errors.Is(err, locationofinterest.ErrNotFound) {
			return 0, 0, forecast.ErrLocationNotFound
		}
		return 0, 0, err
	}
	return m.Latitude(), m.Longitude(), nil
}
