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
	HouseholdId  uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_weather_household"`
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

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:           m.id,
		TenantId:     m.tenantID,
		HouseholdId:  m.householdID,
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
