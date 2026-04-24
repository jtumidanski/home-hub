package householdpreference

import (
	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

// getByIDProvider returns a household preference by primary key.
// Tenant filtering is automatic via GORM callbacks when db.WithContext(ctx) is used.
func getByIDProvider(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

// byTenantUserHouseholdProvider returns a household preference for the composite
// (tenant, user, household) key.
func byTenantUserHouseholdProvider(tenantID, userID, householdID uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ? AND user_id = ? AND household_id = ?", tenantID, userID, householdID)
	})
}
