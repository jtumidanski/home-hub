package main

import (
	"os"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/apps/svc-reminders/reminder"
	"github.com/jtumidanski/home-hub/apps/svc-reminders/user"
	"github.com/jtumidanski/home-hub/packages/shared-go/database"
	"github.com/jtumidanski/home-hub/packages/shared-go/logger"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/jtumidanski/home-hub/packages/shared-go/service"
	"github.com/jtumidanski/home-hub/packages/shared-go/tracing"
	"github.com/sirupsen/logrus"
)

const serviceName = "svc-reminders"

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
	l.Infoln("Starting main service.")

	tdm := service.GetTeardownManager()

	tc, err := tracing.InitTracer(l)(serviceName)
	if err != nil {
		l.WithError(err).Fatal("Unable to initialize tracer.")
	}

	db := database.Connect(l, database.SetMigrations(Migration()))

	// Get users service base URL from environment
	usersServiceURL := os.Getenv("SVC_USERS_BASE_URL")
	if usersServiceURL == "" {
		usersServiceURL = "http://hh-users:8080" // Default for docker-compose
	}

	// Create users service client for resolving user IDs
	usersClient := user.NewClient(usersServiceURL)

	// Start reminder sweeper in background goroutine
	go reminder.StartSweeper(db, l, tdm.Context())

	// Create route initializer (auth is handled by nginx/gateway)
	// We call the users service to resolve the email to user ID
	routeInitializer := func(router *mux.Router, logger logrus.FieldLogger) {
		// Apply user resolver middleware to all routes
		router.Use(user.UserResolverMiddleware(l, usersClient))

		// Initialize reminder routes
		reminder.InitializeRoutes(GetServer())(db)(router, logger)
	}

	server.CreateService(l, tdm.Context(), tdm.WaitGroup(), GetServer().GetPrefix(),
		routeInitializer,
	)

	tdm.TeardownFunc(tracing.Teardown(l)(tc))

	tdm.Wait()
	l.Infoln("Service shutdown.")
}
