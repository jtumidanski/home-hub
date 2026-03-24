package preference

import (
	"context"

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
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByUserProvider(userID uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByUser(userID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) FindOrCreate(tenantID, userID uuid.UUID) (Model, error) {
	m, err := p.ByUserProvider(userID)()
	if err == nil {
		return m, nil
	}

	e, err := create(p.db.WithContext(p.ctx), tenantID, userID)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) UpdateTheme(id uuid.UUID, theme string) (Model, error) {
	e, err := updateTheme(p.db.WithContext(p.ctx), id, theme)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) SetActiveHousehold(id uuid.UUID, householdID uuid.UUID) (Model, error) {
	e, err := setActiveHousehold(p.db.WithContext(p.ctx), id, householdID)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
