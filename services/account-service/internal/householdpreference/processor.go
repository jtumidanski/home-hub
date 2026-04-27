package householdpreference

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByIDProvider(id)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByTenantUserHouseholdProvider(tenantID, userID, householdID uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(byTenantUserHouseholdProvider(tenantID, userID, householdID)(p.db.WithContext(p.ctx)))
}

// FindOrCreate returns the row for (tenant, user, household) or inserts a new
// one with DefaultDashboardID = nil if none exists.
func (p *Processor) FindOrCreate(tenantID, userID, householdID uuid.UUID) (Model, error) {
	m, err := p.ByTenantUserHouseholdProvider(tenantID, userID, householdID)()
	if err == nil {
		return m, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Model{}, err
	}
	e, err := insert(p.db.WithContext(p.ctx), tenantID, userID, householdID)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

// SetDefaultDashboard writes the default dashboard pointer; a nil argument
// clears the field (writes NULL).
func (p *Processor) SetDefaultDashboard(id uuid.UUID, defaultDashboardID *uuid.UUID) (Model, error) {
	e, err := updateFields(p.db.WithContext(p.ctx), id, defaultDashboardID)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
