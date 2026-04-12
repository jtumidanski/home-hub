package retention

import (
	"context"
	"time"

	"github.com/gorilla/mux"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Setup builds the productivity-service reaper, registers the internal HTTP
// endpoints, starts the background loop, and returns once both are wired.
// The caller passes the cancellation context for the loop.
func Setup(ctx context.Context, l logrus.FieldLogger, db *gorm.DB, router *mux.Router, accountURL, internalToken string, interval time.Duration) (*sr.Reaper, error) {
	if err := sr.MigrateRuns(db); err != nil {
		return nil, err
	}
	pc := sr.NewPolicyClient(accountURL, internalToken)
	metrics := sr.NewMetrics("productivity-service")
	reaper := sr.New("productivity-service", db, pc, metrics, l,
		CompletedTasks{},
		DeletedTasksRestoreWindow{},
		AuditTrim{},
	)
	sr.MountInternalEndpoints(router, reaper, internalToken, l)
	router.Handle("/metrics", sr.Handler())

	go sr.Loop(ctx, l, interval, reaper.RunTick)
	return reaper, nil
}
