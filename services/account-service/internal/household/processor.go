package household

import (
	"context"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
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

func (p *Processor) AllProvider() model.Provider[[]Model] {
	return model.SliceMap(Make)(getAll()(p.db.WithContext(p.ctx)))
}

func (p *Processor) Create(tenantID uuid.UUID, name, timezone, units string) (Model, error) {
	e, err := create(p.db.WithContext(p.ctx), tenantID, name, timezone, units)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) CreateWithOwner(tenantID, userID uuid.UUID, name, timezone, units string) (Model, error) {
	m, err := p.Create(tenantID, name, timezone, units)
	if err != nil {
		return Model{}, err
	}

	memProc := membership.NewProcessor(p.l, p.ctx, p.db)
	_, err = memProc.Create(tenantID, m.Id(), userID, "owner")
	if err != nil {
		p.l.WithError(err).Error("Failed to create owner membership for household")
		return Model{}, err
	}

	return m, nil
}

func (p *Processor) Update(id uuid.UUID, name, timezone, units string, latitude, longitude *float64, locationName *string) (Model, error) {
	e, err := update(p.db.WithContext(p.ctx), id, name, timezone, units, latitude, longitude, locationName)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
