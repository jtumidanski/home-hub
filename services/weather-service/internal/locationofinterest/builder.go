package locationofinterest

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTenantIDRequired    = errors.New("tenant ID is required")
	ErrHouseholdIDRequired = errors.New("household ID is required")
	ErrPlaceNameRequired   = errors.New("place name is required")
	ErrLabelTooLong        = errors.New("label must not exceed 64 characters")
	ErrLatitudeOutOfRange  = errors.New("latitude must be between -90 and 90")
	ErrLongitudeOutOfRange = errors.New("longitude must be between -180 and 180")
)

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	label       *string
	placeName   string
	latitude    float64
	longitude   float64
	createdAt   time.Time
	updatedAt   time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder    { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder { b.householdID = id; return b }
func (b *Builder) SetLabel(label *string) *Builder      { b.label = label; return b }
func (b *Builder) SetPlaceName(name string) *Builder    { b.placeName = name; return b }
func (b *Builder) SetLatitude(lat float64) *Builder     { b.latitude = lat; return b }
func (b *Builder) SetLongitude(lon float64) *Builder    { b.longitude = lon; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder    { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder    { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.tenantID == uuid.Nil {
		return Model{}, ErrTenantIDRequired
	}
	if b.householdID == uuid.Nil {
		return Model{}, ErrHouseholdIDRequired
	}
	if strings.TrimSpace(b.placeName) == "" {
		return Model{}, ErrPlaceNameRequired
	}
	if b.label != nil {
		trimmed := strings.TrimSpace(*b.label)
		if len(trimmed) > 64 {
			return Model{}, ErrLabelTooLong
		}
		if trimmed == "" {
			b.label = nil
		} else {
			b.label = &trimmed
		}
	}
	if b.latitude < -90 || b.latitude > 90 {
		return Model{}, ErrLatitudeOutOfRange
	}
	if b.longitude < -180 || b.longitude > 180 {
		return Model{}, ErrLongitudeOutOfRange
	}
	return Model{
		id:          b.id,
		tenantID:    b.tenantID,
		householdID: b.householdID,
		label:       b.label,
		placeName:   b.placeName,
		latitude:    b.latitude,
		longitude:   b.longitude,
		createdAt:   b.createdAt,
		updatedAt:   b.updatedAt,
	}, nil
}
