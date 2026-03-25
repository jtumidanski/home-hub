package externalidentity

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

func (p *Processor) FindByProviderSubject(provider, subject string) model.Provider[Model] {
	return model.Map(Make)(getByProviderAndSubject(provider, subject)(p.db.WithContext(p.ctx)))
}

func (p *Processor) Create(userID uuid.UUID, provider, subject string) (Model, error) {
	e, err := create(p.db.WithContext(p.ctx), userID, provider, subject)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
