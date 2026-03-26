package oauthstate

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrStateNotFound = errors.New("oauth state not found")
	ErrStateExpired  = errors.New("oauth state has expired")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) noTenantDB() *gorm.DB {
	return p.db.WithContext(database.WithoutTenantFilter(p.ctx))
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, redirectURI string) (Model, error) {
	e, err := create(p.noTenantDB(), tenantID, householdID, userID, redirectURI)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) ValidateAndConsume(stateID uuid.UUID) (Model, error) {
	m, err := model.Map(Make)(getByID(stateID)(p.noTenantDB()))()
	if err != nil {
		return Model{}, ErrStateNotFound
	}
	if m.IsExpired() {
		_ = deleteByID(p.noTenantDB(), stateID)
		return Model{}, ErrStateExpired
	}
	_ = deleteByID(p.noTenantDB(), stateID)
	return m, nil
}

func (p *Processor) CleanupExpired() {
	if err := deleteExpired(p.noTenantDB()); err != nil {
		p.l.WithError(err).Warn("failed to cleanup expired oauth states")
	}
}
