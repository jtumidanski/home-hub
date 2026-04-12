// Package retention wires the productivity-service into the shared retention
// framework. It implements two CategoryHandlers: completed-task aging and
// soft-deleted task restore-window expiry. Both run inside a single
// transaction per scope so the cascade (task → task_restorations) is atomic.
package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task/restoration"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/gorm"
)

// CompletedTasks reaps tasks whose completed_at is older than the configured
// window. Cascades to task_restorations rows referencing the deleted tasks.
type CompletedTasks struct{}

func (CompletedTasks) Category() sr.Category { return sr.CatProductivityCompletedTasks }

func (CompletedTasks) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId    uuid.UUID
		HouseholdId uuid.UUID
	}
	var rows []row
	if err := db.Table("tasks").
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

func (CompletedTasks) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	var ids []string
	if err := tx.Table("tasks").
		Where("tenant_id = ? AND household_id = ? AND completed_at IS NOT NULL AND completed_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &ids).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(ids) == 0 {
		return sr.ReapResult{}, nil
	}

	deleted, err := cascadeDeleteTasks(tx, ids)
	if err != nil {
		return sr.ReapResult{}, err
	}
	return sr.ReapResult{Scanned: len(ids), Deleted: deleted}, nil
}

// DeletedTasksRestoreWindow reaps soft-deleted tasks whose deleted_at is
// older than the restore window. Same cascade.
type DeletedTasksRestoreWindow struct{}

func (DeletedTasksRestoreWindow) Category() sr.Category {
	return sr.CatProductivityDeletedTasksRestoreWindow
}

func (DeletedTasksRestoreWindow) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return CompletedTasks{}.DiscoverScopes(ctx, db)
}

func (DeletedTasksRestoreWindow) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	var ids []string
	if err := tx.Table("tasks").
		Where("tenant_id = ? AND household_id = ? AND deleted_at IS NOT NULL AND deleted_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &ids).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(ids) == 0 {
		return sr.ReapResult{}, nil
	}

	deleted, err := cascadeDeleteTasks(tx, ids)
	if err != nil {
		return sr.ReapResult{}, err
	}
	return sr.ReapResult{Scanned: len(ids), Deleted: deleted}, nil
}

// cascadeDeleteTasks removes the listed task ids and their dependent rows
// (task_restorations). Returns the total number of rows removed across both
// tables.
func cascadeDeleteTasks(tx *gorm.DB, ids []string) (int, error) {
	var total int

	r := tx.Where("task_id IN ?", ids).Delete(&restoration.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("id IN ?", ids).Delete(&task.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	return total, nil
}

// AuditTrim reaps old retention_runs rows. This is the system.retention_audit
// category that every reaper-owning service implements against its own table.
type AuditTrim struct{}

func (AuditTrim) Category() sr.Category { return sr.CatSystemRetentionAudit }

func (AuditTrim) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
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
	seenTenant := make(map[uuid.UUID]bool)
	for _, r := range rows {
		if seenTenant[r.TenantId] {
			continue
		}
		seenTenant[r.TenantId] = true
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
