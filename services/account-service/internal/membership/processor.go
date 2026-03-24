package membership

import (
	"context"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Processor struct {
	l   *logrus.Logger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l *logrus.Logger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByUserAndTenantProvider(userID, tenantID uuid.UUID) model.Provider[[]Model] {
	return model.SliceMap(Make)(getByUserAndTenant(userID, tenantID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByHouseholdAndUserProvider(householdID, userID uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByHouseholdAndUser(householdID, userID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, role string) (Model, error) {
	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, userID, role)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) UpdateRole(id uuid.UUID, role string) (Model, error) {
	e, err := updateRole(p.db.WithContext(p.ctx), id, role)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	return deleteByID(p.db.WithContext(p.ctx), id)
}
