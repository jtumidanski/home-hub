package household

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

func (p *Processor) ByTenantIDProvider(tenantID uuid.UUID) model.Provider[[]Model] {
	return model.SliceMap(Make)(getByTenantID(tenantID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) Create(tenantID uuid.UUID, name, timezone, units string) (Model, error) {
	e, err := create(p.db.WithContext(p.ctx), tenantID, name, timezone, units)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(id uuid.UUID, name, timezone, units string) (Model, error) {
	e, err := update(p.db.WithContext(p.ctx), id, name, timezone, units)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
