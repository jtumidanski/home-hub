package event

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrRangeTooLarge = errors.New("query range exceeds 90 days")
)

const maxQueryRange = 90 * 24 * time.Hour

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByID(id uuid.UUID) (Model, error) {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))()
}

func (p *Processor) QueryByHouseholdAndTimeRange(householdID uuid.UUID, start, end time.Time) ([]Model, error) {
	if end.Sub(start) > maxQueryRange {
		return nil, ErrRangeTooLarge
	}
	return model.SliceMap(Make)(getVisibleByHouseholdAndTimeRange(householdID, start, end)(p.db.WithContext(p.ctx)))()
}

func (p *Processor) Upsert(e Entity) error {
	return upsert(p.noTenantDB(), e)
}

func (p *Processor) DeleteBySourceAndExternalIDs(sourceID uuid.UUID, externalIDs []string) error {
	return deleteBySourceAndExternalIDs(p.noTenantDB(), sourceID, externalIDs)
}

func (p *Processor) DeleteByConnection(connectionID uuid.UUID) error {
	return deleteByConnection(p.noTenantDB(), connectionID)
}

func (p *Processor) DeleteBySource(sourceID uuid.UUID) error {
	return deleteBySource(p.noTenantDB(), sourceID)
}

func (p *Processor) CountByConnection(connectionID uuid.UUID) (int64, error) {
	return countByConnection(p.noTenantDB(), connectionID)
}

func (p *Processor) noTenantDB() *gorm.DB {
	return p.db.WithContext(database.WithoutTenantFilter(p.ctx))
}
