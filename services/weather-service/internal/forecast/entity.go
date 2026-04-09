package forecast

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JSONCurrentData CurrentData

func (j JSONCurrentData) Value() (driver.Value, error) { return json.Marshal(j) }
func (j *JSONCurrentData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}
	return json.Unmarshal(b, j)
}

type JSONForecastData []DailyForecast

func (j JSONForecastData) Value() (driver.Value, error) { return json.Marshal(j) }
func (j *JSONForecastData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}
	return json.Unmarshal(b, j)
}

type Entity struct {
	Id           uuid.UUID        `gorm:"type:uuid;primaryKey"`
	TenantId     uuid.UUID        `gorm:"type:uuid;not null;index"`
	HouseholdId  uuid.UUID        `gorm:"type:uuid;not null;index"`
	LocationId   *uuid.UUID       `gorm:"type:uuid;column:location_id"`
	Latitude     float64          `gorm:"type:double precision;not null"`
	Longitude    float64          `gorm:"type:double precision;not null"`
	Units        string           `gorm:"type:text;not null"`
	CurrentData  JSONCurrentData  `gorm:"type:jsonb;not null"`
	ForecastData JSONForecastData `gorm:"type:jsonb;not null"`
	FetchedAt    time.Time        `gorm:"not null"`
	CreatedAt    time.Time        `gorm:"not null"`
	UpdatedAt    time.Time        `gorm:"not null"`
}

func (Entity) TableName() string { return "weather_caches" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	return PostMigration(db)
}

// PostMigration runs idempotent SQL that GORM struct tags cannot express:
// drops the legacy single-column unique index, creates two partial unique
// indexes (one for primary cache rows, one for saved-location rows), and
// adds the FK to locations_of_interest. Safe to re-run.
func PostMigration(db *gorm.DB) error {
	stmts := []string{
		`DROP INDEX IF EXISTS idx_weather_household`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_weather_household_primary ON weather_caches (household_id) WHERE location_id IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_weather_household_location ON weather_caches (household_id, location_id) WHERE location_id IS NOT NULL`,
		`DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_weather_location'
  ) THEN
    ALTER TABLE weather_caches
      ADD CONSTRAINT fk_weather_location
      FOREIGN KEY (location_id)
      REFERENCES locations_of_interest(id)
      ON DELETE CASCADE;
  END IF;
END $$;`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:           m.id,
		TenantId:     m.tenantID,
		HouseholdId:  m.householdID,
		LocationId:   m.locationID,
		Latitude:     m.latitude,
		Longitude:    m.longitude,
		Units:        m.units,
		CurrentData:  JSONCurrentData(m.currentData),
		ForecastData: JSONForecastData(m.forecastData),
		FetchedAt:    m.fetchedAt,
		CreatedAt:    m.createdAt,
		UpdatedAt:    m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetLocationID(e.LocationId).
		SetLatitude(e.Latitude).
		SetLongitude(e.Longitude).
		SetUnits(e.Units).
		SetCurrentData(CurrentData(e.CurrentData)).
		SetForecastData([]DailyForecast(e.ForecastData)).
		SetFetchedAt(e.FetchedAt).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
