// Package retention wires tracker-service into the shared retention framework.
package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/gorm"
)

func discoverUserScopes(db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId uuid.UUID
		UserId   uuid.UUID
	}
	var rows []row
	if err := db.Table("tracking_items").
		Select("DISTINCT tenant_id, user_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]sr.Scope, 0, len(rows))
	for _, r := range rows {
		out = append(out, sr.Scope{TenantId: r.TenantId, Kind: sr.ScopeUser, ScopeId: r.UserId})
	}
	return out, nil
}

// Entries reaps tracking_entries by date. No upward cascade — the parent
// tracking_item is preserved.
type Entries struct{}

func (Entries) Category() sr.Category { return sr.CatTrackerEntries }

func (Entries) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverUserScopes(db)
}

func (Entries) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	r := tx.Where("tenant_id = ? AND user_id = ? AND date < ?", scope.TenantId, scope.ScopeId, cutoff).
		Delete(&entry.Entity{})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	return sr.ReapResult{Scanned: int(r.RowsAffected), Deleted: int(r.RowsAffected)}, nil
}

// DeletedItemsRestoreWindow reaps soft-deleted tracking_items past their
// restore window. Cascades to all tracking_entries for the item.
type DeletedItemsRestoreWindow struct{}

func (DeletedItemsRestoreWindow) Category() sr.Category {
	return sr.CatTrackerDeletedItemsRestoreWindow
}

func (DeletedItemsRestoreWindow) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverUserScopes(db)
}

func (DeletedItemsRestoreWindow) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	var ids []uuid.UUID
	if err := tx.Table("tracking_items").
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NOT NULL AND deleted_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &ids).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(ids) == 0 {
		return sr.ReapResult{}, nil
	}

	var total int
	r := tx.Where("tracking_item_id IN ?", ids).Delete(&entry.Entity{})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	total += int(r.RowsAffected)
	r = tx.Where("id IN ?", ids).Delete(&trackingitem.Entity{})
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
	// Tracker is user-scoped, but the audit category is household-scoped per
	// the registry. Discover from retention_runs directly so we don't need a
	// mapping table.
	type row struct {
		TenantId  uuid.UUID
		ScopeId   uuid.UUID
		ScopeKind string
	}
	var rows []row
	if err := db.Table("retention_runs").
		Select("DISTINCT tenant_id, scope_id, scope_kind").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]sr.Scope, 0, len(rows))
	seen := make(map[uuid.UUID]bool)
	for _, r := range rows {
		if seen[r.TenantId] {
			continue
		}
		seen[r.TenantId] = true
		out = append(out, sr.Scope{TenantId: r.TenantId, Kind: sr.ScopeHousehold, ScopeId: r.ScopeId})
	}
	return out, nil
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
