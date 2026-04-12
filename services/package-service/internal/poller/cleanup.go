package poller

import (
	"context"
	"time"

	"github.com/jtumidanski/home-hub/services/package-service/internal/tracking"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

// CleanupConfig holds configuration for the residual cleanup job. After
// task-036, archive transitions and hard deletion live in the retention
// framework; this loop only handles the stale-marking pass which is unrelated
// to retention windows.
type CleanupConfig struct {
	StaleAfterDays int
}

// StartCleanupLoop runs the residual stale-marking job. Archive and hard
// delete are now driven by the retention framework.
func (e *Engine) StartCleanupLoop(ctx context.Context, cfg CleanupConfig) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	e.l.WithField("stale_after_days", cfg.StaleAfterDays).Info("package stale-marking loop started")

	for {
		select {
		case <-ctx.Done():
			e.l.Info("package stale-marking loop stopped")
			return
		case <-ticker.C:
			e.runStaleMark(ctx, cfg.StaleAfterDays)
		}
	}
}

func (e *Engine) runStaleMark(ctx context.Context, staleDays int) {
	noTenantCtx := database.WithoutTenantFilter(ctx)
	db := e.db.WithContext(noTenantCtx)
	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(staleDays) * 24 * time.Hour)

	result := db.Model(&tracking.Entity{}).
		Where("status IN ? AND last_status_change_at IS NOT NULL AND last_status_change_at < ?",
			[]string{tracking.StatusPreTransit, tracking.StatusInTransit, tracking.StatusOutForDelivery},
			cutoff).
		Updates(map[string]interface{}{
			"status":     tracking.StatusStale,
			"updated_at": now,
		})

	if result.Error != nil {
		e.l.WithError(result.Error).Error("failed to mark stale packages")
		return
	}
	if result.RowsAffected > 0 {
		e.l.WithField("count", result.RowsAffected).Info("marked packages as stale")
	}
}

// silence unused-import in case Engine is touched.
var _ = (*gorm.DB)(nil)
