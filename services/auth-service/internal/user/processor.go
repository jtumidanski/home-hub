package user

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
	return byIDProvider(id)(p.db.WithContext(p.ctx))
}

func (p *Processor) ByEmailProvider(email string) model.Provider[Model] {
	return byEmailProvider(email)(p.db.WithContext(p.ctx))
}

func (p *Processor) ByIDsProvider(ids []uuid.UUID) model.Provider[[]Model] {
	return model.SliceMap(modelFromEntity)(getByIDs(ids)(p.db.WithContext(p.ctx)))
}

func (p *Processor) FindOrCreate(email, displayName, givenName, familyName, avatarURL string) (Model, error) {
	m, err := p.ByEmailProvider(email)()
	if err == nil {
		return m, nil
	}

	e, err := create(p.db.WithContext(p.ctx), email, displayName, givenName, familyName, avatarURL)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
