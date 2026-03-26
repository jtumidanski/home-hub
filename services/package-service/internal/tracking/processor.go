package tracking

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/package-service/internal/carrier"
	"github.com/jtumidanski/home-hub/services/package-service/internal/trackingevent"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const refreshCooldown = 5 * time.Minute

var (
	ErrNotFound       = errors.New("package not found")
	ErrDuplicate      = errors.New("tracking number already exists in this household")
	ErrHouseholdLimit = errors.New("household has reached the maximum number of active packages")
	ErrNotOwner       = errors.New("only the package creator can perform this action")
	ErrRefreshTooSoon = errors.New("package was recently refreshed")
	ErrNotArchived    = errors.New("package is not archived")
)

type CreateAttrs struct {
	TrackingNumber string
	Carrier        string
	Label          string
	Notes          string
	Private        bool
}

type UpdateAttrs struct {
	Label   *string
	Notes   *string
	Carrier *string
	Private *bool
}

type Processor struct {
	l         logrus.FieldLogger
	ctx       context.Context
	db        *gorm.DB
	maxActive int
	carriers  *carrier.Registry
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, maxActive int, carriers *carrier.Registry) *Processor {
	return &Processor{l: l, ctx: ctx, db: db, maxActive: maxActive, carriers: carriers}
}

func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, attrs CreateAttrs) (Model, error) {
	if _, err := NewBuilder().SetTrackingNumber(attrs.TrackingNumber).SetCarrier(attrs.Carrier).Build(); err != nil {
		return Model{}, err
	}

	exists, err := existsByHouseholdAndTrackingNumber(p.db.WithContext(p.ctx), householdID, attrs.TrackingNumber)
	if err != nil {
		return Model{}, err
	}
	if exists {
		return Model{}, ErrDuplicate
	}

	count, err := countActiveByHousehold(p.db.WithContext(p.ctx), householdID)
	if err != nil {
		return Model{}, err
	}
	if count >= int64(p.maxActive) {
		return Model{}, ErrHouseholdLimit
	}

	now := time.Now().UTC()
	e := Entity{
		Id:                 uuid.New(),
		TenantId:           tenantID,
		HouseholdId:        householdID,
		UserId:             userID,
		TrackingNumber:     attrs.TrackingNumber,
		Carrier:            attrs.Carrier,
		Status:             StatusPreTransit,
		Private:            attrs.Private,
		LastStatusChangeAt: &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if attrs.Label != "" {
		e.Label = &attrs.Label
	}
	if attrs.Notes != "" {
		e.Notes = &attrs.Notes
	}

	if err := create(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	// Initial poll — best effort, don't fail creation if carrier is unavailable
	p.pollEntity(&e)

	return Make(e)
}

func (p *Processor) Get(id uuid.UUID) (Model, error) {
	m, err := p.ByIDProvider(id)()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return m, nil
}

func (p *Processor) GetTrackingEvents(packageID uuid.UUID) ([]trackingevent.Model, error) {
	entities, err := trackingevent.GetByPackageID(packageID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	models := make([]trackingevent.Model, 0, len(entities))
	for _, e := range entities {
		m, merr := trackingevent.Make(e)
		if merr != nil {
			return nil, merr
		}
		models = append(models, m)
	}
	return models, nil
}

func (p *Processor) List(householdID uuid.UUID, includeArchived bool, filterStatuses []string, hasETA bool, sortField string) ([]Model, error) {
	var entities []Entity
	var err error

	if hasETA {
		statuses := filterStatuses
		if len(statuses) == 0 {
			statuses = []string{StatusPreTransit, StatusInTransit, StatusOutForDelivery}
		}
		entities, err = getByHouseholdWithETA(householdID, statuses)(p.db.WithContext(p.ctx))()
	} else if includeArchived {
		entities, err = getByHouseholdWithArchived(householdID)(p.db.WithContext(p.ctx))()
	} else {
		statuses := filterStatuses
		if len(statuses) == 0 {
			statuses = []string{StatusPreTransit, StatusInTransit, StatusOutForDelivery, StatusDelivered, StatusException, StatusStale}
		}
		entities, err = getByHouseholdAndStatus(householdID, statuses)(p.db.WithContext(p.ctx))()
	}

	if err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(entities))
	for _, e := range entities {
		m, merr := Make(e)
		if merr != nil {
			return nil, merr
		}
		models = append(models, m)
	}
	return models, nil
}

func (p *Processor) Update(id, userID uuid.UUID, attrs UpdateAttrs) (Model, error) {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.UserId != userID {
		return Model{}, ErrNotOwner
	}

	if attrs.Label != nil {
		e.Label = attrs.Label
	}
	if attrs.Notes != nil {
		e.Notes = attrs.Notes
	}
	if attrs.Carrier != nil {
		if !validCarriers[*attrs.Carrier] {
			return Model{}, ErrInvalidCarrier
		}
		e.Carrier = *attrs.Carrier
	}
	if attrs.Private != nil {
		e.Private = *attrs.Private
	}
	e.UpdatedAt = time.Now().UTC()

	if err := update(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id, userID uuid.UUID) error {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return ErrNotFound
	}
	if e.UserId != userID {
		return ErrNotOwner
	}
	return deleteByID(p.db.WithContext(p.ctx), id)
}

func (p *Processor) Archive(id uuid.UUID) (Model, error) {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}

	now := time.Now().UTC()
	e.Status = StatusArchived
	e.ArchivedAt = &now
	e.UpdatedAt = now

	if err := update(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Unarchive(id uuid.UUID) (Model, error) {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Status != StatusArchived {
		return Model{}, ErrNotArchived
	}

	now := time.Now().UTC()
	e.Status = StatusDelivered
	e.ArchivedAt = nil
	e.UpdatedAt = now

	if err := update(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Refresh(id uuid.UUID) (Model, error) {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}

	if e.LastPolledAt != nil && time.Since(*e.LastPolledAt) < refreshCooldown {
		return Model{}, ErrRefreshTooSoon
	}

	p.pollEntity(&e)
	return Make(e)
}

func (p *Processor) Summary(householdID uuid.UUID) (SummaryResult, error) {
	var result SummaryResult

	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	db := p.db.WithContext(p.ctx)

	var err error
	result.ArrivingTodayCount, err = countArrivingToday(db, householdID, today, tomorrow)
	if err != nil {
		return result, err
	}

	result.InTransitCount, err = countInTransit(db, householdID)
	if err != nil {
		return result, err
	}

	result.ExceptionCount, err = countExceptions(db, householdID)
	if err != nil {
		return result, err
	}

	return result, nil
}

// PollEntity queries the carrier API for updated tracking data and persists the results.
// Exported for use by the background polling scheduler.
func (p *Processor) PollEntity(e *Entity) {
	p.pollEntity(e)
}

func (p *Processor) pollEntity(e *Entity) {
	if p.carriers == nil {
		return
	}

	client, ok := p.carriers.Get(e.Carrier)
	if !ok {
		p.l.WithField("carrier", e.Carrier).Warn("no carrier client registered")
		return
	}

	ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
	defer cancel()

	result, err := client.Track(ctx, e.TrackingNumber)
	if err != nil {
		p.l.WithError(err).WithField("carrier", e.Carrier).Warn("carrier tracking poll failed")
		return
	}

	now := time.Now().UTC()
	e.LastPolledAt = &now
	e.UpdatedAt = now

	if result.Found {
		if result.Status != "" && result.Status != e.Status {
			e.Status = result.Status
			e.LastStatusChangeAt = &now
		}
		if result.EstimatedDelivery != nil {
			e.EstimatedDelivery = result.EstimatedDelivery
		}
		if result.ActualDelivery != nil {
			e.ActualDelivery = result.ActualDelivery
		}
	}

	// Use a context without tenant filter for the update since this may be called
	// from a background goroutine without tenant context.
	noTenantCtx := database.WithoutTenantFilter(p.ctx)
	if err := update(p.db.WithContext(noTenantCtx), e); err != nil {
		p.l.WithError(err).Error("failed to update package after poll")
		return
	}

	// Store tracking events
	for _, ev := range result.Events {
		loc := ev.Location
		var locPtr *string
		if loc != "" {
			locPtr = &loc
		}
		raw := ev.RawStatus
		var rawPtr *string
		if raw != "" {
			rawPtr = &raw
		}
		if err := trackingevent.Create(p.db.WithContext(noTenantCtx), e.Id, ev.Timestamp, ev.Status, ev.Description, locPtr, rawPtr); err != nil {
			p.l.WithError(err).Warn("failed to store tracking event")
		}
	}
}
