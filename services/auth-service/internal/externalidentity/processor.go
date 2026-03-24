package externalidentity

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

func (p *Processor) FindByProviderSubject(provider, subject string) model.Provider[Entity] {
	return getByProviderAndSubject(provider, subject)(p.db.WithContext(p.ctx))
}

func (p *Processor) Create(userID uuid.UUID, provider, subject string) (Entity, error) {
	return create(p.db.WithContext(p.ctx), userID, provider, subject)
}
