package dashboard

import (
	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

// visibleToCaller returns all dashboards for (tenant, household) either
// household-scoped or owned by the caller user.
func visibleToCaller(tenantID, householdID, callerUserID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ? AND household_id = ? AND (user_id IS NULL OR user_id = ?)",
			tenantID, householdID, callerUserID).
			Order("sort_order ASC, created_at ASC")
	})
}

func householdScoped(tenantID, householdID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ? AND household_id = ? AND user_id IS NULL", tenantID, householdID).
			Order("sort_order ASC, created_at ASC")
	})
}

func maxSortOrderInScope(db *gorm.DB, tenantID, householdID uuid.UUID, userID *uuid.UUID) (int, error) {
	var result struct{ Max *int }
	q := db.Model(&Entity{}).Select("MAX(sort_order) AS max").
		Where("tenant_id = ? AND household_id = ?", tenantID, householdID)
	if userID == nil {
		q = q.Where("user_id IS NULL")
	} else {
		q = q.Where("user_id = ?", *userID)
	}
	if err := q.Scan(&result).Error; err != nil {
		return 0, err
	}
	if result.Max == nil {
		return -1, nil
	}
	return *result.Max, nil
}

func countHouseholdScoped(db *gorm.DB, tenantID, householdID uuid.UUID) (int64, error) {
	var n int64
	err := db.Model(&Entity{}).
		Where("tenant_id = ? AND household_id = ? AND user_id IS NULL", tenantID, householdID).
		Count(&n).Error
	return n, err
}
