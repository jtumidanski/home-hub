package forecast

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTenantIDRequired    = errors.New("tenant ID is required")
	ErrHouseholdIDRequired = errors.New("household ID is required")
	ErrLatitudeOutOfRange  = errors.New("latitude must be between -90 and 90")
	ErrLongitudeOutOfRange = errors.New("longitude must be between -180 and 180")
	ErrUnitsRequired       = errors.New("units is required")
	ErrCurrentDataRequired = errors.New("current data is required")
)

type Builder struct {
	id           uuid.UUID
	tenantID     uuid.UUID
	householdID  uuid.UUID
	latitude     float64
	longitude    float64
	units        string
	currentData  CurrentData
	forecastData []DailyForecast
	fetchedAt    time.Time
	createdAt    time.Time
	updatedAt    time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = id
	return b
}

func (b *Builder) SetTenantID(tenantID uuid.UUID) *Builder {
	b.tenantID = tenantID
	return b
}

func (b *Builder) SetHouseholdID(householdID uuid.UUID) *Builder {
	b.householdID = householdID
	return b
}

func (b *Builder) SetLatitude(lat float64) *Builder {
	b.latitude = lat
	return b
}

func (b *Builder) SetLongitude(lon float64) *Builder {
	b.longitude = lon
	return b
}

func (b *Builder) SetUnits(units string) *Builder {
	b.units = units
	return b
}

func (b *Builder) SetCurrentData(data CurrentData) *Builder {
	b.currentData = data
	return b
}

func (b *Builder) SetForecastData(data []DailyForecast) *Builder {
	b.forecastData = data
	return b
}

func (b *Builder) SetFetchedAt(t time.Time) *Builder {
	b.fetchedAt = t
	return b
}

func (b *Builder) SetCreatedAt(t time.Time) *Builder {
	b.createdAt = t
	return b
}

func (b *Builder) SetUpdatedAt(t time.Time) *Builder {
	b.updatedAt = t
	return b
}

func (b *Builder) Build() (Model, error) {
	if b.tenantID == uuid.Nil {
		return Model{}, ErrTenantIDRequired
	}
	if b.householdID == uuid.Nil {
		return Model{}, ErrHouseholdIDRequired
	}
	if b.latitude < -90 || b.latitude > 90 {
		return Model{}, ErrLatitudeOutOfRange
	}
	if b.longitude < -180 || b.longitude > 180 {
		return Model{}, ErrLongitudeOutOfRange
	}
	if b.units == "" {
		return Model{}, ErrUnitsRequired
	}
	return Model{
		id:           b.id,
		tenantID:     b.tenantID,
		householdID:  b.householdID,
		latitude:     b.latitude,
		longitude:    b.longitude,
		units:        b.units,
		currentData:  b.currentData,
		forecastData: b.forecastData,
		fetchedAt:    b.fetchedAt,
		createdAt:    b.createdAt,
		updatedAt:    b.updatedAt,
	}, nil
}
