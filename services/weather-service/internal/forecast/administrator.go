package forecast

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID uuid.UUID, locationID *uuid.UUID, lat, lon float64, units string, current CurrentData, daily []DailyForecast) (Entity, error) {
	now := time.Now().UTC()

	// Upsert based on (household_id, location_id) — partial unique indexes
	// enforce uniqueness for both the primary (NULL) and saved-location cases.
	// We can't use ON CONFLICT against partial indexes portably, so do a
	// look-up + insert/update inside a transaction instead.
	var actual Entity
	err := db.Transaction(func(tx *gorm.DB) error {
		q := tx.Where("household_id = ?", householdID)
		if locationID == nil {
			q = q.Where("location_id IS NULL")
		} else {
			q = q.Where("location_id = ?", *locationID)
		}

		var existing Entity
		err := q.First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			e := Entity{
				Id:           uuid.New(),
				TenantId:     tenantID,
				HouseholdId:  householdID,
				LocationId:   locationID,
				Latitude:     lat,
				Longitude:    lon,
				Units:        units,
				CurrentData:  JSONCurrentData(current),
				ForecastData: JSONForecastData(daily),
				FetchedAt:    now,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := tx.Create(&e).Error; err != nil {
				return err
			}
			actual = e
			return nil
		}
		if err != nil {
			return err
		}

		existing.Latitude = lat
		existing.Longitude = lon
		existing.Units = units
		existing.CurrentData = JSONCurrentData(current)
		existing.ForecastData = JSONForecastData(daily)
		existing.FetchedAt = now
		existing.UpdatedAt = now
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}
		actual = existing
		return nil
	})
	if err != nil {
		return Entity{}, err
	}
	return actual, nil
}

func deleteByHouseholdID(db *gorm.DB, householdID uuid.UUID) error {
	return db.Where("household_id = ? AND location_id IS NULL", householdID).Delete(&Entity{}).Error
}
