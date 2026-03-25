package forecast

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func create(db *gorm.DB, tenantID, householdID uuid.UUID, lat, lon float64, units string, current CurrentData, daily []DailyForecast) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:           uuid.New(),
		TenantId:     tenantID,
		HouseholdId:  householdID,
		Latitude:     lat,
		Longitude:    lon,
		Units:        units,
		CurrentData:  JSONCurrentData(current),
		ForecastData: JSONForecastData(daily),
		FetchedAt:    now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "household_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"latitude", "longitude", "units", "current_data", "forecast_data", "fetched_at", "updated_at",
		}),
	}).Create(&e)

	if result.Error != nil {
		return Entity{}, result.Error
	}

	// Re-read to get the actual row (upsert may have used existing ID)
	var actual Entity
	if err := db.Where("household_id = ?", householdID).First(&actual).Error; err != nil {
		return Entity{}, err
	}
	return actual, nil
}

func deleteByHouseholdID(db *gorm.DB, householdID uuid.UUID) error {
	return db.Where("household_id = ?", householdID).Delete(&Entity{}).Error
}
