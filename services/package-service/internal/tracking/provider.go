package tracking

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

func getByHouseholdAndStatus(householdID uuid.UUID, statuses []string) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ? AND status IN ?", householdID, statuses).
			Order("estimated_delivery ASC NULLS LAST, created_at DESC")
	})
}

func getByHouseholdWithArchived(householdID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ?", householdID).
			Order("estimated_delivery ASC NULLS LAST, created_at DESC")
	})
}

func getByHouseholdWithETA(householdID uuid.UUID, statuses []string) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ? AND status IN ? AND estimated_delivery IS NOT NULL", householdID, statuses).
			Order("estimated_delivery ASC")
	})
}

func countActiveByHousehold(db *gorm.DB, householdID uuid.UUID) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("household_id = ? AND status NOT IN ?", householdID, []string{StatusArchived}).
		Count(&count).Error
	return count, err
}

func existsByHouseholdAndTrackingNumber(db *gorm.DB, householdID uuid.UUID, trackingNumber string) (bool, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("household_id = ? AND tracking_number = ?", householdID, trackingNumber).
		Count(&count).Error
	return count > 0, err
}

func create(db *gorm.DB, e *Entity) error {
	return db.Create(e).Error
}

func update(db *gorm.DB, e *Entity) error {
	return db.Save(e).Error
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}

type SummaryResult struct {
	ArrivingTodayCount int64
	InTransitCount     int64
	ExceptionCount     int64
}
