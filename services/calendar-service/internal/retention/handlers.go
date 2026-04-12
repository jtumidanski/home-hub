// Package retention wires calendar-service into the shared retention framework.
package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/event"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/gorm"
)

func discoverHouseholdScopes(db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId    uuid.UUID
		HouseholdId uuid.UUID
	}
	var rows []row
	if err := db.Table("calendar_events").
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

// PastEvents reaps calendar_events whose end_time is older than the configured
// window. Leaf-level: no children, no cascade. Future events are untouched.
type PastEvents struct{}

func (PastEvents) Category() sr.Category { return sr.CatCalendarPastEvents }

func (PastEvents) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverHouseholdScopes(db)
}

func (PastEvents) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	r := tx.Where("tenant_id = ? AND household_id = ? AND end_time < ?", scope.TenantId, scope.ScopeId, cutoff).
		Delete(&event.Entity{})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	return sr.ReapResult{Scanned: int(r.RowsAffected), Deleted: int(r.RowsAffected)}, nil
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
