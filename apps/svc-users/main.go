package main

import (
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/apps/svc-users/household"
	"github.com/jtumidanski/home-hub/apps/svc-users/user"
	"github.com/jtumidanski/home-hub/apps/svc-users/user/preference"
	"github.com/jtumidanski/home-hub/packages/shared-go/auth"
	"github.com/jtumidanski/home-hub/packages/shared-go/database"
	"github.com/jtumidanski/home-hub/packages/shared-go/logger"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/jtumidanski/home-hub/packages/shared-go/service"
	"github.com/jtumidanski/home-hub/packages/shared-go/tracing"
	"github.com/sirupsen/logrus"
)

const serviceName = "svc-users"

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

	// Create auth providers
	userProvider := auth.NewSimpleUserProvider(db)
	roleProvider := auth.NewSimpleRoleProvider(db)

	// Create optional auth middleware - allows service-to-service calls without auth
	// while still authenticating requests that come through nginx
	authMiddleware := auth.OptionalMiddleware(l, db, userProvider, roleProvider)

	// Create custom route initializer with auth middleware
	authRouteInitializer := func(router *mux.Router, logger logrus.FieldLogger) {
		// Apply optional auth middleware to all routes
		router.Use(authMiddleware)

		// Initialize user, household, and preference routes
		user.InitializeRoutes(GetServer())(db)(router, logger)
		household.InitializeRoutes(GetServer())(db)(router, logger)
		preference.InitializeRoutes(GetServer())(db)(router, logger)
	}

	server.CreateService(l, tdm.Context(), tdm.WaitGroup(), GetServer().GetPrefix(),
		authRouteInitializer,
	)

	tdm.TeardownFunc(tracing.Teardown(l)(tc))

	tdm.Wait()
	l.Infoln("Service shutdown.")
}
