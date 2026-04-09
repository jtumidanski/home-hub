package locationofinterest

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ErrCapReached's message is part of the public API contract — see PRD §4.1
// (task-026). Clients display it verbatim; do not change without updating the
// PRD and the corresponding processor test.
var (
	ErrNotFound   = errors.New("location of interest not found")
	ErrCapReached = errors.New("Households can save up to 10 locations of interest. Remove one to add another.")
	MaxLocations  = 10
)

// CacheWarmer is the minimal slice of forecast.Processor needed to warm the
// cache after a new location is created. Defined here as an interface so the
// package does not import forecast (avoiding an import cycle when forecast
// later imports this package for the locationId resolution path).
type CacheWarmer interface {
	WarmLocationCache(tenantID, householdID uuid.UUID, locationID uuid.UUID, lat, lon float64) error
}

type Processor struct {
	l      logrus.FieldLogger
	ctx    context.Context
	db     *gorm.DB
	warmer CacheWarmer
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, warmer CacheWarmer) *Processor {
	return &Processor{l: l, ctx: ctx, db: db, warmer: warmer}
}

func (p *Processor) List(householdID uuid.UUID) ([]Model, error) {
	entities, err := ListByHousehold(householdID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		models[i] = m
	}
	return models, nil
}

func (p *Processor) Get(householdID, id uuid.UUID) (Model, error) {
	e, err := GetByID(id, householdID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return Make(e)
}

// CreateInput captures fields a caller may set when creating a location.
type CreateInput struct {
	Label     *string
	PlaceName string
	Latitude  float64
	Longitude float64
}

func (p *Processor) Create(tenantID, householdID uuid.UUID, input CreateInput) (Model, error) {
	placeName := strings.TrimSpace(input.PlaceName)
	if placeName == "" {
		return Model{}, ErrPlaceNameRequired
	}
	var label *string
	if input.Label != nil {
		trimmed := strings.TrimSpace(*input.Label)
		if len(trimmed) > 64 {
			return Model{}, ErrLabelTooLong
		}
		if trimmed != "" {
			label = &trimmed
		}
	}

	lat := math.Round(input.Latitude*10000) / 10000
	lon := math.Round(input.Longitude*10000) / 10000

	if _, err := NewBuilder().
		SetTenantID(tenantID).
		SetHouseholdID(householdID).
		SetLabel(label).
		SetPlaceName(placeName).
		SetLatitude(lat).
		SetLongitude(lon).
		Build(); err != nil {
		return Model{}, err
	}

	count, err := countByHousehold(p.db.WithContext(p.ctx), householdID)
	if err != nil {
		return Model{}, err
	}
	if count >= int64(MaxLocations) {
		return Model{}, ErrCapReached
	}

	e := Entity{
		TenantId:    tenantID,
		HouseholdId: householdID,
		Label:       label,
		PlaceName:   placeName,
		Latitude:    lat,
		Longitude:   lon,
	}
	if err := createLocation(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	if p.warmer != nil {
		if err := p.warmer.WarmLocationCache(tenantID, householdID, e.Id, lat, lon); err != nil {
			p.l.WithError(err).WithField("location_id", e.Id.String()).
				Warn("failed to warm cache for new location of interest")
		}
	}

	return Make(e)
}

func (p *Processor) UpdateLabel(householdID, id uuid.UUID, newLabel *string) (Model, error) {
	e, err := GetByID(id, householdID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if newLabel == nil {
		// no-op (label not provided)
		return Make(e)
	}
	trimmed := strings.TrimSpace(*newLabel)
	if len(trimmed) > 64 {
		return Model{}, ErrLabelTooLong
	}
	if trimmed == "" {
		e.Label = nil
	} else {
		e.Label = &trimmed
	}
	if err := updateLocation(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(householdID, id uuid.UUID) error {
	if _, err := GetByID(id, householdID)(p.db.WithContext(p.ctx))(); err != nil {
		return ErrNotFound
	}
	return deleteLocation(p.db.WithContext(p.ctx), id, householdID)
}
