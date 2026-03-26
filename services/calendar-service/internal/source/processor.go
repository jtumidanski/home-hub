package source

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
	ErrNotFound = errors.New("source not found")
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

func (p *Processor) ListByConnection(connectionID uuid.UUID) ([]Model, error) {
	return model.SliceMap(Make)(getByConnection(connectionID)(p.noTenantDB()))()
}

func (p *Processor) CreateOrUpdate(tenantID, householdID, connectionID uuid.UUID, externalID, name string, primary bool, color string) (Model, error) {
	existing, err := model.Map(Make)(getByConnectionAndExternalID(connectionID, externalID)(p.noTenantDB()))()
	if err == nil {
		if err := updateNameAndColor(p.noTenantDB(), existing.Id(), name, color, primary); err != nil {
			return Model{}, err
		}
		return model.Map(Make)(getByID(existing.Id())(p.noTenantDB()))()
	}

	e, err := create(p.noTenantDB(), tenantID, householdID, connectionID, externalID, name, primary, color)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) ToggleVisibility(id uuid.UUID, visible bool) error {
	return updateVisibility(p.db.WithContext(p.ctx), id, visible)
}

func (p *Processor) UpdateSyncToken(id uuid.UUID, syncToken string) error {
	return updateSyncToken(p.noTenantDB(), id, syncToken)
}

func (p *Processor) ClearSyncToken(id uuid.UUID) error {
	return updateSyncToken(p.noTenantDB(), id, "")
}

func (p *Processor) DeleteByConnection(connectionID uuid.UUID) error {
	return deleteByConnection(p.noTenantDB(), connectionID)
}

func (p *Processor) noTenantDB() *gorm.DB {
	return p.db.WithContext(database.WithoutTenantFilter(p.ctx))
}
