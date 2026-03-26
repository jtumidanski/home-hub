package poller

import (
	"context"
	"time"

	"github.com/jtumidanski/home-hub/services/package-service/internal/carrier"
	"github.com/jtumidanski/home-hub/services/package-service/internal/tracking"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Engine handles background polling of carrier APIs for package status updates.
type Engine struct {
	db        *gorm.DB
	carriers  *carrier.Registry
	maxActive int
	l         logrus.FieldLogger
}

// NewEngine creates a new polling engine.
func NewEngine(db *gorm.DB, carriers *carrier.Registry, maxActive int, l logrus.FieldLogger) *Engine {
	return &Engine{db: db, carriers: carriers, maxActive: maxActive, l: l}
}

// StartPollLoop runs the background polling loop at the given intervals.
func (e *Engine) StartPollLoop(ctx context.Context, normalInterval, urgentInterval time.Duration) {
	ticker := time.NewTicker(normalInterval)
	defer ticker.Stop()

	e.l.WithFields(logrus.Fields{
		"normal_interval": normalInterval.String(),
		"urgent_interval": urgentInterval.String(),
	}).Info("package polling loop started")

	for {
		select {
		case <-ctx.Done():
			e.l.Info("package polling loop stopped")
			return
		case <-ticker.C:
			e.pollDue(ctx, normalInterval, urgentInterval)
		}
	}
}

func (e *Engine) pollDue(ctx context.Context, normalInterval, urgentInterval time.Duration) {
	noTenantCtx := database.WithoutTenantFilter(ctx)
	now := time.Now().UTC()

	var packages []tracking.Entity

	// Find packages due for polling based on their status and last poll time
	err := e.db.WithContext(noTenantCtx).
		Where("status IN ? AND (last_polled_at IS NULL OR "+
			"(status = 'out_for_delivery' AND last_polled_at < ?) OR "+
			"(status != 'out_for_delivery' AND last_polled_at < ?))",
			[]string{tracking.StatusPreTransit, tracking.StatusInTransit, tracking.StatusOutForDelivery},
			now.Add(-urgentInterval),
			now.Add(-normalInterval),
		).
		Find(&packages).Error

	if err != nil {
		e.l.WithError(err).Error("failed to query packages for polling")
		return
	}

	if len(packages) == 0 {
		return
	}

	e.l.WithField("count", len(packages)).Info("polling packages")

	for _, pkg := range packages {
		select {
		case <-ctx.Done():
			return
		default:
			p := pkg // copy for closure
			proc := tracking.NewProcessor(e.l, noTenantCtx, e.db, e.maxActive, e.carriers)
			proc.PollEntity(&p)
		}
	}
}
