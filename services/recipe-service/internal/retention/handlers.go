// Package retention wires recipe-service into the shared retention framework.
package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/gorm"
)

func discoverRecipeScopes(db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId    uuid.UUID
		HouseholdId uuid.UUID
	}
	var rows []row
	if err := db.Table("recipes").
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

// DeletedRecipesRestoreWindow reaps soft-deleted recipes past their restore
// window. Cascade: recipes → tags → recipe_restorations → plan_items.
type DeletedRecipesRestoreWindow struct{}

func (DeletedRecipesRestoreWindow) Category() sr.Category {
	return sr.CatRecipeDeletedRecipesRestoreWindow
}

func (DeletedRecipesRestoreWindow) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverRecipeScopes(db)
}

func (DeletedRecipesRestoreWindow) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	var ids []uuid.UUID
	if err := tx.Table("recipes").
		Where("tenant_id = ? AND household_id = ? AND deleted_at IS NOT NULL AND deleted_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &ids).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(ids) == 0 {
		return sr.ReapResult{}, nil
	}

	deleted, err := cascadeDeleteRecipes(tx, ids)
	if err != nil {
		return sr.ReapResult{}, err
	}
	return sr.ReapResult{Scanned: len(ids), Deleted: deleted}, nil
}

func cascadeDeleteRecipes(tx *gorm.DB, ids []uuid.UUID) (int, error) {
	var total int

	r := tx.Where("recipe_id IN ?", ids).Delete(&recipe.TagEntity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("recipe_id IN ?", ids).Delete(&normalization.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("recipe_id IN ?", ids).Delete(&recipe.RestorationEntity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("recipe_id IN ?", ids).Delete(&planitem.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("id IN ?", ids).Delete(&recipe.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)
	return total, nil
}

// RestorationAudit trims recipe_restorations rows older than the audit window.
type RestorationAudit struct{}

func (RestorationAudit) Category() sr.Category { return sr.CatRecipeRestorationAudit }

func (RestorationAudit) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverRecipeScopes(db)
}

func (RestorationAudit) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	// Restoration rows in this service do not carry tenant_id; they are linked
	// via recipe_id. Scope by joining on the recipes table.
	r := tx.Exec(`
		DELETE FROM recipe_restorations
		WHERE recipe_id IN (
			SELECT id FROM recipes
			WHERE tenant_id = ? AND household_id = ?
		) AND restored_at < ?`, scope.TenantId, scope.ScopeId, cutoff)
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	return sr.ReapResult{Scanned: int(r.RowsAffected), Deleted: int(r.RowsAffected)}, nil
}

// AuditTrim reaps the service's own retention_runs rows.
type AuditTrim struct{}

func (AuditTrim) Category() sr.Category { return sr.CatSystemRetentionAudit }

func (AuditTrim) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return discoverRecipeScopes(db)
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
