// Package retention wires package-service into the shared retention framework.
// It replaces the env-var-driven cleanup loop in internal/poller/cleanup.go
// with policy-driven reapers that read window values from account-service.
//
// Two categories live here:
//
//   - package.archive_window — delivered packages older than the window are
//     transitioned to the archived status (a soft-delete equivalent). This
//     does NOT delete rows; it preserves the audit trail before the second
//     window kicks in.
//   - package.archived_delete_window — archived packages older than the
//     window are hard-deleted along with their tracking_events.
//
// Stale-marking (the third stage of the old loop) is unrelated to retention
// and remains in internal/poller/cleanup.go for now.
package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/package-service/internal/tracking"
	"github.com/jtumidanski/home-hub/services/package-service/internal/trackingevent"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/gorm"
)

func discoverHouseholdScopes(db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId    uuid.UUID
		HouseholdId uuid.UUID
	}
	var rows []row
	if err := db.Table("packages").
		Select("DISTINCT tenant_id, household_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]sr.Scope, 0, len(rows))
	for _, r := range rows {
		out = append(out, sr.Scope{TenantId: r.TenantId, Kind: sr.ScopeHousehold, ScopeId: r.HouseholdId})
	}
	return out, nil
}

// ArchiveWindow transitions delivered packages older than the window into the
// archived state. It does not delete rows. The Reaped count reflects the
// number of packages whose status was flipped.
type ArchiveWindow struct{}

func (ArchiveWindow) Category() sr.Category { return sr.CatPackageArchiveWindow }

func (ArchiveWindow) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverHouseholdScopes(db)
}

func (ArchiveWindow) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(days) * 24 * time.Hour)

	r := tx.Model(&tracking.Entity{}).
		Where("tenant_id = ? AND household_id = ? AND status = ? AND last_status_change_at IS NOT NULL AND last_status_change_at < ?",
			scope.TenantId, scope.ScopeId, tracking.StatusDelivered, cutoff).
		Updates(map[string]interface{}{
			"status":      tracking.StatusArchived,
			"archived_at": now,
			"updated_at":  now,
		})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	// Treat scanned == deleted == rows-affected here. The audit semantics for
	// "archive transitions" match: how many rows did this category move out
	// of the live state.
	return sr.ReapResult{Scanned: int(r.RowsAffected), Deleted: int(r.RowsAffected)}, nil
}

// ArchivedDeleteWindow hard-deletes archived packages whose archived_at is
// older than the configured window. Cascades to tracking_events.
type ArchivedDeleteWindow struct{}

func (ArchivedDeleteWindow) Category() sr.Category { return sr.CatPackageArchivedDeleteWindow }

func (ArchivedDeleteWindow) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverHouseholdScopes(db)
}

func (ArchivedDeleteWindow) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	var ids []uuid.UUID
	if err := tx.Table("packages").
		Where("tenant_id = ? AND household_id = ? AND status = ? AND archived_at IS NOT NULL AND archived_at < ?",
			scope.TenantId, scope.ScopeId, tracking.StatusArchived, cutoff).
		Pluck("id", &ids).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(ids) == 0 {
		return sr.ReapResult{}, nil
	}

	var total int
	r := tx.Where("package_id IN ?", ids).Delete(&trackingevent.Entity{})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("id IN ?", ids).Delete(&tracking.Entity{})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	total += int(r.RowsAffected)
	return sr.ReapResult{Scanned: len(ids), Deleted: total}, nil
}

// AuditTrim trims this service's retention_runs.
type AuditTrim struct{}

func (AuditTrim) Category() sr.Category { return sr.CatSystemRetentionAudit }

func (AuditTrim) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverHouseholdScopes(db)
}

func (AuditTrim) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	r := tx.Where("tenant_id = ? AND started_at < ?", scope.TenantId, cutoff).
		Delete(&sr.RunEntity{})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	return sr.ReapResult{Scanned: int(r.RowsAffected), Deleted: int(r.RowsAffected)}, nil
}
