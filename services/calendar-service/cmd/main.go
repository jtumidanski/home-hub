package main

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/config"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/connection"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/crypto"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/event"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/oauthstate"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/source"
	calendarsync "github.com/jtumidanski/home-hub/services/calendar-service/internal/sync"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "calendar-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB,
		database.SetMigrations(
			oauthstate.Migration,
			connection.Migration,
			source.Migration,
			event.Migration,
		),
	)

	enc, err := crypto.NewEncryptor(cfg.TokenEncryptionKey)
	if err != nil {
		log.Fatalf("failed to initialize token encryptor: %v", err)
	}

	gcClient := googlecal.NewClient(cfg.GoogleCalendarClientID, cfg.GoogleCalendarSecret, l)
	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	syncEngine := calendarsync.NewEngine(db, gcClient, enc, l)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go syncEngine.StartLoop(ctx, cfg.SyncInterval)

	syncTrigger := func(conn connection.Model) {
		syncEngine.SyncConnection(conn)
	}

	cascadeDelete := func(ctx context.Context, connectionID uuid.UUID) {
		syncEngine.DeleteConnectionData(ctx, connectionID)
	}

	ownerCheck := func(db *gorm.DB, l logrus.FieldLogger, r *http.Request, connID uuid.UUID) (uuid.UUID, error) {
		proc := connection.NewProcessor(l, r.Context(), db)
		conn, err := proc.ByIDProvider(connID)()
		if err != nil {
			return uuid.Nil, err
		}
		return conn.UserID(), nil
	}

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			connection.InitializeRoutes(db, gcClient, enc, cfg, syncTrigger, cascadeDelete)(l, si, api)
			source.InitializeRoutes(db, ownerCheck)(l, si, api)
			event.InitializeRoutes(db)(l, si, api)
		}).
		Run()
}
