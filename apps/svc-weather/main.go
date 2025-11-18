package main

import (
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/apps/svc-weather/cache"
	"github.com/jtumidanski/home-hub/apps/svc-weather/geokey"
	"github.com/jtumidanski/home-hub/apps/svc-weather/household"
	"github.com/jtumidanski/home-hub/apps/svc-weather/openmeteo"
	"github.com/jtumidanski/home-hub/apps/svc-weather/scheduler"
	"github.com/jtumidanski/home-hub/apps/svc-weather/weather"
	"github.com/jtumidanski/home-hub/packages/shared-go/health"
	"github.com/jtumidanski/home-hub/packages/shared-go/logger"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/jtumidanski/home-hub/packages/shared-go/service"
	"github.com/jtumidanski/home-hub/packages/shared-go/tracing"
	"github.com/sirupsen/logrus"
)

const (
	serviceName = "svc-weather"
	version     = "1.0.0"
)

type Server struct {
	baseUrl string
	prefix  string
}

func (s Server) GetBaseURL() string {
	return s.baseUrl
}

func (s Server) GetPrefix() string {
	return s.prefix
}

func GetServer() Server {
	return Server{
		baseUrl: "",
		prefix:  "/api",
	}
}

func main() {
	l := logger.CreateLogger(serviceName)
	l.Infoln("Starting weather service")

	// Load configuration
	cfg := LoadConfig()
	l.WithFields(logrus.Fields{
		"redis_url":         cfg.RedisURL,
		"current_ttl":       cfg.CurrentTTL,
		"forecast_ttl":      cfg.ForecastTTL,
		"geohash_precision": cfg.GeohashPrecision,
	}).Info("Configuration loaded")

	tdm := service.GetTeardownManager()

	tc, err := tracing.InitTracer(l)(serviceName)
	if err != nil {
		l.WithError(err).Fatal("Unable to initialize tracer")
	}

	// Initialize cache client (Redis or in-memory fallback)
	var cacheClient cache.Client
	if cfg.RedisURL != "" {
		redisClient, err := cache.NewRedisClient(cfg.RedisURL, l)
		if err != nil {
			l.WithError(err).Warn("Failed to connect to Redis, falling back to in-memory cache")
			cacheClient = cache.NewMemoryClient(l)
		} else {
			cacheClient = redisClient
			tdm.TeardownFunc(func() {
				if err := redisClient.Close(); err != nil {
					l.WithError(err).Error("Failed to close Redis client")
				}
			})
		}
	} else {
		cacheClient = cache.NewMemoryClient(l)
	}

	// Initialize household resolver
	householdResolver := household.NewHTTPResolver(
		cfg.SvcUsersBaseURL,
		cfg.SvcUsersTimeout,
		cfg.HouseholdCacheTTL,
		l,
	)

	// Initialize Open-Meteo client
	meteoClient := openmeteo.NewHTTPClient(
		cfg.OpenMeteoBaseURL,
		cfg.OpenMeteoTimeout,
		l,
	)

	// Initialize geokey generator and key builder
	geoGen := geokey.NewGenerator(cfg.GeohashPrecision)
	keyBuilder := geokey.NewKeyBuilder()

	// Initialize weather provider
	weatherProvider := weather.NewCacheProvider(
		cacheClient,
		householdResolver,
		meteoClient,
		geoGen,
		keyBuilder,
		cfg.CurrentTTL,
		cfg.ForecastTTL,
		cfg.StaleMax,
		l,
	)

	// Initialize background scheduler
	weatherScheduler := scheduler.NewWeatherScheduler(
		weatherProvider.RefreshCurrent,
		weatherProvider.RefreshForecast,
		cfg.CurrentTTL,
		cfg.ForecastTTL,
		cfg.RefreshJitter,
		l,
	)

	// Start scheduler
	if err := weatherScheduler.Start(tdm.Context()); err != nil {
		l.WithError(err).Fatal("Failed to start weather scheduler")
	}

	// Add scheduler shutdown to teardown
	tdm.TeardownFunc(func() {
		l.Info("Shutting down weather scheduler")
		if err := weatherScheduler.Stop(); err != nil {
			l.WithError(err).Error("Error stopping weather scheduler")
		}
	})

	// Create health check aggregator with cache check
	healthChecker := health.NewAggregator(
		health.NewCacheCheck(cacheClient),
	)

	// Create route initializer (auth is handled by nginx/gateway)
	routeInitializer := func(router *mux.Router, logger logrus.FieldLogger) {
		// Register health endpoint (unauthenticated)
		router.HandleFunc("/health", health.Handler(serviceName, version, healthChecker, logger)).Methods("GET")

		// Initialize weather routes (pass scheduler for household tracking)
		weather.InitializeRoutes(GetServer(), weatherProvider, weatherScheduler)(router, logger)
	}

	server.CreateService(l, tdm.Context(), tdm.WaitGroup(), GetServer().GetPrefix(),
		routeInitializer,
	)

	tdm.TeardownFunc(tracing.Teardown(l)(tc))

	tdm.Wait()
	l.Infoln("Service shutdown")
}
