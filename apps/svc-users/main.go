package main

import (
	"github.com/jtumidanski/home-hub/packages/shared-go/database"
	"github.com/jtumidanski/home-hub/packages/shared-go/logger"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/jtumidanski/home-hub/packages/shared-go/service"
	"github.com/jtumidanski/home-hub/packages/shared-go/tracing"
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

	_ = database.Connect(l, database.SetMigrations())

	server.CreateService(l, tdm.Context(), tdm.WaitGroup(), GetServer().GetPrefix())

	tdm.TeardownFunc(tracing.Teardown(l)(tc))

	tdm.Wait()
	l.Infoln("Service shutdown.")
}
