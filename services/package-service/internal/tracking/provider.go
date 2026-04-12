package tracking

import (
	"time"

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

func countActiveByHousehold(householdID uuid.UUID) func(db *gorm.DB) (int64, error) {
	return func(db *gorm.DB) (int64, error) {
		var count int64
		err := db.Model(&Entity{}).
			Where("household_id = ? AND status NOT IN ?", householdID, []string{StatusArchived}).
			Count(&count).Error
		return count, err
	}
}

func existsByHouseholdAndTrackingNumber(householdID uuid.UUID, trackingNumber string) func(db *gorm.DB) (bool, error) {
	return func(db *gorm.DB) (bool, error) {
		var count int64
		err := db.Model(&Entity{}).
			Where("household_id = ? AND tracking_number = ?", householdID, trackingNumber).
			Count(&count).Error
		return count > 0, err
	}
}

func countArrivingToday(householdID uuid.UUID, today, tomorrow time.Time) func(db *gorm.DB) (int64, error) {
	return func(db *gorm.DB) (int64, error) {
		var count int64
		err := db.Model(&Entity{}).
			Where("household_id = ? AND estimated_delivery >= ? AND estimated_delivery < ? AND status IN ?",
				householdID, today, tomorrow, []string{StatusPreTransit, StatusInTransit, StatusOutForDelivery}).
			Count(&count).Error
		return count, err
	}
}

func countInTransit(householdID uuid.UUID) func(db *gorm.DB) (int64, error) {
	return func(db *gorm.DB) (int64, error) {
		var count int64
		err := db.Model(&Entity{}).
			Where("household_id = ? AND status IN ?", householdID,
				[]string{StatusPreTransit, StatusInTransit, StatusOutForDelivery}).
			Count(&count).Error
		return count, err
	}
}

func countExceptions(householdID uuid.UUID) func(db *gorm.DB) (int64, error) {
	return func(db *gorm.DB) (int64, error) {
		var count int64
		err := db.Model(&Entity{}).
			Where("household_id = ? AND status = ?", householdID, StatusException).
			Count(&count).Error
		return count, err
	}
}

type SummaryResult struct {
	ArrivingTodayCount int64
	InTransitCount     int64
	ExceptionCount     int64
}
