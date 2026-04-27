// Package retention wires dashboard-service into the shared retention framework.
//
// v1 never auto-purges dashboards (Defaults[CatDashboardDashboards] = 0); the
// plumbing exists to surface the category in the UI and the retention_runs
// audit table, but Reap is a no-op by design. If operators later configure a
// positive day-count for this category, the reaper will still short-circuit
// until the handler grows real purge semantics.
//
// AuditTrim trims this service's rows from the shared retention_runs table
// using the standard system.retention_audit category; every retention-using
// service registers it.
package retention

import (
	"context"

	"github.com/google/uuid"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/gorm"
	"time"
)

func discoverHouseholdScopes(db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId    uuid.UUID
		HouseholdId uuid.UUID
	}
	var rows []row
	if err := db.Table("dashboards").
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

// DashboardsRetention is the dashboards category. v1 never auto-purges rows —
// Reap is intentionally a no-op. The plumbing exists so the category surfaces
// in the retention UI and audit log.
type DashboardsRetention struct{}

func (DashboardsRetention) Category() sr.Category { return sr.CatDashboardDashboards }

func (DashboardsRetention) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverHouseholdScopes(db)
}

func (DashboardsRetention) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	return sr.ReapResult{}, nil
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
