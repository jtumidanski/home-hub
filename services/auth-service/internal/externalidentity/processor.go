package externalidentity

import (
	"context"
	"time"

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

// FindByProviderSubject looks up an external identity by provider and subject.
func (p *Processor) FindByProviderSubject(provider, subject string) model.Provider[Entity] {
	return getByProviderAndSubject(provider, subject)(p.db.WithContext(p.ctx))
}

// Create creates a new external identity mapping.
func (p *Processor) Create(userID uuid.UUID, provider, subject string) (Entity, error) {
	now := time.Now().UTC()
	e := &Entity{
		Id:              uuid.New(),
		UserId:          userID,
		Provider:        provider,
		ProviderSubject: subject,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := p.db.WithContext(p.ctx).Create(e).Error; err != nil {
		return Entity{}, err
	}
	return *e, nil
}
