package household

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired           = errors.New("household name is required")
	ErrTimezoneRequired       = errors.New("household timezone is required")
	ErrUnitsRequired          = errors.New("household units is required")
	ErrPartialCoordinates     = errors.New("both latitude and longitude must be provided together")
	ErrLatitudeOutOfRange     = errors.New("latitude must be between -90 and 90")
	ErrLongitudeOutOfRange    = errors.New("longitude must be between -180 and 180")
)

type Builder struct {
	id           uuid.UUID
	tenantID     uuid.UUID
	name         string
	timezone     string
	units        string
	latitude     *float64
	longitude    *float64
	locationName *string
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

func (b *Builder) SetName(name string) *Builder {
	b.name = name
	return b
}

func (b *Builder) SetTimezone(timezone string) *Builder {
	b.timezone = timezone
	return b
}

func (b *Builder) SetUnits(units string) *Builder {
	b.units = units
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

func (b *Builder) SetLatitude(lat *float64) *Builder {
	b.latitude = lat
	return b
}

func (b *Builder) SetLongitude(lon *float64) *Builder {
	b.longitude = lon
	return b
}

func (b *Builder) SetLocationName(name *string) *Builder {
	b.locationName = name
	return b
}

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if b.timezone == "" {
		return Model{}, ErrTimezoneRequired
	}
	if b.units == "" {
		return Model{}, ErrUnitsRequired
	}
	if (b.latitude == nil) != (b.longitude == nil) {
		return Model{}, ErrPartialCoordinates
	}
	if b.latitude != nil && (*b.latitude < -90 || *b.latitude > 90) {
		return Model{}, ErrLatitudeOutOfRange
	}
	if b.longitude != nil && (*b.longitude < -180 || *b.longitude > 180) {
		return Model{}, ErrLongitudeOutOfRange
	}
	return Model{
		id:           b.id,
		tenantID:     b.tenantID,
		name:         b.name,
		timezone:     b.timezone,
		units:        b.units,
		latitude:     b.latitude,
		longitude:    b.longitude,
		locationName: b.locationName,
		createdAt:    b.createdAt,
		updatedAt:    b.updatedAt,
	}, nil
}
