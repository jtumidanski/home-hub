package plan

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

type ListFilters struct {
	StartsOn *time.Time
	Page     int
	PageSize int
}

func getAll(filters ListFilters) func(db *gorm.DB) ([]Entity, int64, error) {
	return func(db *gorm.DB) ([]Entity, int64, error) {
		query := db.Model(&Entity{})

		if filters.StartsOn != nil {
			query = query.Where("starts_on = ?", *filters.StartsOn)
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (filters.Page - 1) * filters.PageSize
		var entities []Entity
		err := query.
			Order("starts_on DESC").
			Offset(offset).
			Limit(filters.PageSize).
			Find(&entities).Error
		return entities, total, err
	}
}

func getByHouseholdAndStartsOn(db *gorm.DB, householdID uuid.UUID, startsOn time.Time) (Entity, error) {
	var e Entity
	err := db.Where("household_id = ? AND starts_on = ?", householdID, startsOn).First(&e).Error
	return e, err
}
