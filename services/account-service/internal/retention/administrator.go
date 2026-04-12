package retention

import (
	"time"

	"github.com/google/uuid"
	sharedretention "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/gorm"
)

func upsertOverride(db *gorm.DB, tenantID uuid.UUID, scopeKind sharedretention.ScopeKind, scopeID uuid.UUID, cat sharedretention.Category, days int) (Entity, error) {
	now := time.Now().UTC()
	var existing Entity
	err := db.Where("tenant_id = ? AND scope_kind = ? AND scope_id = ? AND category = ?",
		tenantID, string(scopeKind), scopeID, string(cat)).First(&existing).Error
	if err == nil {
		existing.RetentionDays = days
		existing.UpdatedAt = now
		if err := db.Save(&existing).Error; err != nil {
			return Entity{}, err
		}
		return existing, nil
	}
	e := Entity{
		Id:            uuid.New(),
		TenantId:      tenantID,
		ScopeKind:     string(scopeKind),
		ScopeId:       scopeID,
		Category:      string(cat),
		RetentionDays: days,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func deleteOverride(db *gorm.DB, tenantID uuid.UUID, scopeKind sharedretention.ScopeKind, scopeID uuid.UUID, cat sharedretention.Category) error {
	return db.Where("tenant_id = ? AND scope_kind = ? AND scope_id = ? AND category = ?",
		tenantID, string(scopeKind), scopeID, string(cat)).Delete(&Entity{}).Error
}

func listOverrides(db *gorm.DB, tenantID uuid.UUID, scopeKind sharedretention.ScopeKind, scopeID uuid.UUID) ([]Entity, error) {
	var rows []Entity
	err := db.Where("tenant_id = ? AND scope_kind = ? AND scope_id = ?",
		tenantID, string(scopeKind), scopeID).Find(&rows).Error
	return rows, err
}
