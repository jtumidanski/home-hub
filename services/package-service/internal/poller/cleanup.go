package poller

import (
	"context"
	"time"

	"github.com/jtumidanski/home-hub/services/package-service/internal/tracking"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CleanupConfig holds configuration for the cleanup job.
type CleanupConfig struct {
	ArchiveAfterDays int
	DeleteAfterDays  int
	StaleAfterDays   int
}

// StartCleanupLoop runs the daily cleanup job.
func (e *Engine) StartCleanupLoop(ctx context.Context, cfg CleanupConfig) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	e.l.WithFields(logrus.Fields{
		"archive_after_days": cfg.ArchiveAfterDays,
		"delete_after_days":  cfg.DeleteAfterDays,
		"stale_after_days":   cfg.StaleAfterDays,
	}).Info("package cleanup loop started")

	for {
		select {
		case <-ctx.Done():
			e.l.Info("package cleanup loop stopped")
			return
		case <-ticker.C:
			e.runCleanup(ctx, cfg)
		}
	}
}

func (e *Engine) runCleanup(ctx context.Context, cfg CleanupConfig) {
	noTenantCtx := database.WithoutTenantFilter(ctx)
	db := e.db.WithContext(noTenantCtx)
	now := time.Now().UTC()

	// 1. Mark stale: packages with no status change for N days
	e.markStale(db, now, cfg.StaleAfterDays)

	// 2. Auto-archive: delivered packages after N days
	e.autoArchive(db, now, cfg.ArchiveAfterDays)

	// 3. Hard delete: archived packages after N days
	e.hardDelete(db, now, cfg.DeleteAfterDays)
}

func (e *Engine) markStale(db *gorm.DB, now time.Time, staleDays int) {
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
	} else if result.RowsAffected > 0 {
		e.l.WithField("count", result.RowsAffected).Info("marked packages as stale")
	}
}

func (e *Engine) autoArchive(db *gorm.DB, now time.Time, archiveDays int) {
	cutoff := now.Add(-time.Duration(archiveDays) * 24 * time.Hour)

	result := db.Model(&tracking.Entity{}).
		Where("status = ? AND last_status_change_at IS NOT NULL AND last_status_change_at < ?",
			tracking.StatusDelivered, cutoff).
		Updates(map[string]interface{}{
			"status":      tracking.StatusArchived,
			"archived_at": now,
			"updated_at":  now,
		})

	if result.Error != nil {
		e.l.WithError(result.Error).Error("failed to auto-archive delivered packages")
	} else if result.RowsAffected > 0 {
		e.l.WithField("count", result.RowsAffected).Info("auto-archived delivered packages")
	}
}

func (e *Engine) hardDelete(db *gorm.DB, now time.Time, deleteDays int) {
	cutoff := now.Add(-time.Duration(deleteDays) * 24 * time.Hour)

	result := db.Where("status = ? AND archived_at IS NOT NULL AND archived_at < ?",
		tracking.StatusArchived, cutoff).
		Delete(&tracking.Entity{})

	if result.Error != nil {
		e.l.WithError(result.Error).Error("failed to hard-delete archived packages")
	} else if result.RowsAffected > 0 {
		e.l.WithField("count", result.RowsAffected).Info("hard-deleted old archived packages")
	}
}
