package main

import (
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/apps/svc-meals/ai"
	"github.com/jtumidanski/home-hub/apps/svc-meals/meal"
	"github.com/jtumidanski/home-hub/apps/svc-meals/user"
	"github.com/jtumidanski/home-hub/packages/shared-go/database"
	"github.com/jtumidanski/home-hub/packages/shared-go/health"
	"github.com/jtumidanski/home-hub/packages/shared-go/logger"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/jtumidanski/home-hub/packages/shared-go/service"
	"github.com/jtumidanski/home-hub/packages/shared-go/tracing"
	"github.com/sirupsen/logrus"
)

const (
	serviceName = "svc-meals"
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
		prefix:  "/api/",
	}
}

func main() {
	l := logger.CreateLogger(serviceName)
	l.Infoln("Starting meals service.")

	tdm := service.GetTeardownManager()

	tc, err := tracing.InitTracer(l)(serviceName)
	if err != nil {
		l.WithError(err).Fatal("Unable to initialize tracer.")
	}

	db := database.Connect(l, database.SetMigrations(Migration()))

	// Create health check aggregator
	healthChecker := health.NewAggregator(
		health.NewDatabaseCheck(db),
	)

	// Get AI service base URL from environment
	aiServiceURL := os.Getenv("SVC_AI_BASE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "http://hh-ai:8080" // Default for docker-compose
	}

	// Get AI service timeout from environment (default 120 seconds for recipe parsing)
	aiTimeoutSeconds := 120
	if timeoutStr := os.Getenv("SVC_AI_TIMEOUT_SECONDS"); timeoutStr != "" {
		if parsed, err := strconv.Atoi(timeoutStr); err == nil && parsed > 0 {
			aiTimeoutSeconds = parsed
		}
	}
	aiTimeout := time.Duration(aiTimeoutSeconds) * time.Second

	// Create AI service client
	aiClient := ai.NewClient(aiServiceURL, aiTimeout, l)

	// Get users service base URL from environment
	usersServiceURL := os.Getenv("SVC_USERS_BASE_URL")
	if usersServiceURL == "" {
		usersServiceURL = "http://hh-users:8080" // Default for docker-compose
	}

	// Create users service client for resolving user IDs
	usersClient := user.NewClient(usersServiceURL)

	// Create route initializer
	routeInitializer := func(router *mux.Router, logger logrus.FieldLogger) {
		// Register health endpoint (unauthenticated)
		router.HandleFunc("/health", health.Handler(serviceName, version, healthChecker, logger)).Methods("GET")

		// Apply user resolver middleware to all routes
		router.Use(user.UserResolverMiddleware(l, usersClient))

		// Initialize meal routes
		meal.InitializeRoutes(GetServer(), aiClient)(db)(router, logger)
	}

	// Create server with extended timeouts for AI recipe parsing
	server.New(l).
		WithContext(tdm.Context()).
		WithWaitGroup(tdm.WaitGroup()).
		SetBasePath(GetServer().GetPrefix()).
		SetReadTimeout(200 * time.Second).  // Long read timeout for large recipe text
		SetWriteTimeout(200 * time.Second). // Long write timeout for AI processing responses
		SetRouteInitializers(routeInitializer).
		Run()

	tdm.TeardownFunc(tracing.Teardown(l)(tc))

	tdm.Wait()
	l.Infoln("Service shutdown.")
}
