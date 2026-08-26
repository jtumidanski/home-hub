package retention

import (
	"context"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	sr "github.com/jtumidanski/home-hub/shared/go/retention"
)

func Setup(ctx context.Context, l logrus.FieldLogger, db *gorm.DB, router *mux.Router, accountURL, internalToken string, interval time.Duration) (*sr.Reaper, error) {
	if err := sr.MigrateRuns(db); err != nil {
		return nil, err
	}
	pc := sr.NewPolicyClient(accountURL, internalToken)
	metrics := sr.NewMetrics("package-service")
	reaper := sr.New("package-service", db, pc, metrics, l,
		ArchiveWindow{},
		ArchivedDeleteWindow{},
		AuditTrim{},
	)
	sr.MountInternalEndpoints(router, reaper, internalToken, l)
	router.Handle("/metrics", sr.Handler())
	go sr.Loop(ctx, l, interval, reaper.RunTick)
	return reaper, nil
}
