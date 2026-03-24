package main

import (
	"context"

	"github.com/jtumidanski/home-hub/services/auth-service/internal/config"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/externalidentity"
	authjwt "github.com/jtumidanski/home-hub/services/auth-service/internal/jwt"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/oidcprovider"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/refreshtoken"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/resource"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/user"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "auth-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			user.Migration,
			externalidentity.Migration,
			oidcprovider.Migration,
			refreshtoken.Migration,
		),
	)

	issuer, err := authjwt.NewIssuer(cfg.JWTPrivateKey, cfg.JWTKeyID)
	if err != nil {
		l.WithError(err).Fatal("failed to initialize JWT issuer")
	}

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(resource.InitializeRoutes(l, db, issuer)).
		Run()
}
