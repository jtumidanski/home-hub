// Package retention wires workout-service into the shared retention framework.
package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/region"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/theme"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/gorm"
)

func discoverUserScopes(db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId uuid.UUID
		UserId   uuid.UUID
	}
	var rows []row
	if err := db.Table("themes").
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

// Performances reaps old performance + performance_set rows. The "performed
// at" timestamp is sourced from performances.created_at as a pragmatic proxy
// since the actual performance date is derived via week + day_of_week.
type Performances struct{}

func (Performances) Category() sr.Category { return sr.CatWorkoutPerformances }

func (Performances) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverUserScopes(db)
}

func (Performances) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	var ids []uuid.UUID
	if err := tx.Table("performances").
		Where("tenant_id = ? AND user_id = ? AND created_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &ids).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(ids) == 0 {
		return sr.ReapResult{}, nil
	}

	var total int
	r := tx.Where("performance_id IN ?", ids).Delete(&performance.SetEntity{})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("id IN ?", ids).Delete(&performance.Entity{})
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	total += int(r.RowsAffected)
	return sr.ReapResult{Scanned: len(ids), Deleted: total}, nil
}

// DeletedCatalogRestoreWindow reaps soft-deleted themes/regions/exercises
// past their restore window. Cascade order: theme → regions → exercises →
// performances → performance_sets, all in one transaction per call.
type DeletedCatalogRestoreWindow struct{}

func (DeletedCatalogRestoreWindow) Category() sr.Category {
	return sr.CatWorkoutDeletedCatalogRestoreWindow
}

func (DeletedCatalogRestoreWindow) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverUserScopes(db)
}

func (DeletedCatalogRestoreWindow) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	scanned := 0
	deleted := 0

	// Step 1: themes whose deleted_at is past cutoff. Take their IDs and
	// expand region IDs that belong to them (whether soft-deleted or not, so
	// the cascade leaves nothing dangling).
	var themeIDs []uuid.UUID
	if err := tx.Table("themes").
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NOT NULL AND deleted_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &themeIDs).Error; err != nil {
		return sr.ReapResult{}, err
	}

	var regionIDs []uuid.UUID
	if len(themeIDs) > 0 {
		// All exercises that belong to one of the doomed themes are doomed too.
		var exerciseIDsFromThemes []uuid.UUID
		if err := tx.Table("exercises").
			Where("tenant_id = ? AND user_id = ? AND theme_id IN ?", scope.TenantId, scope.ScopeId, themeIDs).
			Pluck("id", &exerciseIDsFromThemes).Error; err != nil {
			return sr.ReapResult{}, err
		}
		dN, err := cascadeDeleteExercises(tx, scope, exerciseIDsFromThemes)
		if err != nil {
			return sr.ReapResult{}, err
		}
		deleted += dN
		scanned += len(exerciseIDsFromThemes)
	}

	// Step 2: regions whose deleted_at is past cutoff (independent of themes).
	if err := tx.Table("regions").
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NOT NULL AND deleted_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &regionIDs).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(regionIDs) > 0 {
		var exerciseIDsFromRegions []uuid.UUID
		if err := tx.Table("exercises").
			Where("tenant_id = ? AND user_id = ? AND region_id IN ?", scope.TenantId, scope.ScopeId, regionIDs).
			Pluck("id", &exerciseIDsFromRegions).Error; err != nil {
			return sr.ReapResult{}, err
		}
		dN, err := cascadeDeleteExercises(tx, scope, exerciseIDsFromRegions)
		if err != nil {
			return sr.ReapResult{}, err
		}
		deleted += dN
		scanned += len(exerciseIDsFromRegions)
	}

	// Step 3: standalone soft-deleted exercises.
	var standaloneExerciseIDs []uuid.UUID
	if err := tx.Table("exercises").
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NOT NULL AND deleted_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &standaloneExerciseIDs).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(standaloneExerciseIDs) > 0 {
		dN, err := cascadeDeleteExercises(tx, scope, standaloneExerciseIDs)
		if err != nil {
			return sr.ReapResult{}, err
		}
		deleted += dN
		scanned += len(standaloneExerciseIDs)
	}

	// Step 4: now delete the now-empty regions and themes themselves.
	if len(regionIDs) > 0 {
		r := tx.Where("id IN ?", regionIDs).Delete(&region.Entity{})
		if r.Error != nil {
			return sr.ReapResult{}, r.Error
		}
		deleted += int(r.RowsAffected)
		scanned += int(r.RowsAffected)
	}
	if len(themeIDs) > 0 {
		r := tx.Where("id IN ?", themeIDs).Delete(&theme.Entity{})
		if r.Error != nil {
			return sr.ReapResult{}, r.Error
		}
		deleted += int(r.RowsAffected)
		scanned += int(r.RowsAffected)
	}

	return sr.ReapResult{Scanned: scanned, Deleted: deleted}, nil
}

func cascadeDeleteExercises(tx *gorm.DB, scope sr.Scope, exerciseIDs []uuid.UUID) (int, error) {
	if len(exerciseIDs) == 0 {
		return 0, nil
	}
	var deleted int

	// performances → performance_sets
	var perfIDs []uuid.UUID
	if err := tx.Table("performances").
		Where("tenant_id = ? AND user_id = ?", scope.TenantId, scope.ScopeId).
		Where("planned_item_id IN (SELECT id FROM planned_items WHERE exercise_id IN ?)", exerciseIDs).
		Pluck("id", &perfIDs).Error; err != nil {
		return 0, err
	}
	if len(perfIDs) > 0 {
		r := tx.Where("performance_id IN ?", perfIDs).Delete(&performance.SetEntity{})
		if r.Error != nil {
			return 0, r.Error
		}
		deleted += int(r.RowsAffected)
		r = tx.Where("id IN ?", perfIDs).Delete(&performance.Entity{})
		if r.Error != nil {
			return 0, r.Error
		}
		deleted += int(r.RowsAffected)
	}

	// Finally the exercises themselves.
	r := tx.Where("id IN ?", exerciseIDs).Delete(&exercise.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	deleted += int(r.RowsAffected)
	return deleted, nil
}

// AuditTrim trims this service's retention_runs.
type AuditTrim struct{}

func (AuditTrim) Category() sr.Category { return sr.CatSystemRetentionAudit }

func (AuditTrim) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId uuid.UUID
		ScopeId  uuid.UUID
	}
	var rows []row
	if err := db.Table("retention_runs").
		Select("DISTINCT tenant_id, scope_id").
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
