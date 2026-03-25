package task

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("task not found")
	ErrNotDeleted    = errors.New("task is not deleted")
	ErrRestoreWindow = errors.New("restore window expired")
)

const restoreWindowDays = 3

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

func (p *Processor) AllProvider(includeDeleted bool) model.Provider[[]Model] {
	return model.SliceMap(Make)(getAll(includeDeleted)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByStatusProvider(status string) model.Provider[[]Model] {
	return model.SliceMap(Make)(getByStatus(status)(p.db.WithContext(p.ctx)))
}

func (p *Processor) Create(tenantID, householdID uuid.UUID, title, notes string, dueOn *time.Time, rolloverEnabled bool) (Model, error) {
	if _, err := NewBuilder().SetTitle(title).Build(); err != nil {
		return Model{}, err
	}
	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, title, notes, "pending", dueOn, rolloverEnabled)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(id uuid.UUID, title, notes, status string, dueOn *time.Time, rolloverEnabled bool, userID uuid.UUID) (Model, error) {
	e, err := update(p.db.WithContext(p.ctx), id, title, notes, status, dueOn, rolloverEnabled, userID)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	return softDelete(p.db.WithContext(p.ctx), id)
}

func (p *Processor) Restore(id uuid.UUID) error {
	m, err := p.ByIDProvider(id)()
	if err != nil {
		return ErrNotFound
	}
	if !m.IsDeleted() {
		return ErrNotDeleted
	}
	if time.Since(*m.DeletedAt()) > restoreWindowDays*24*time.Hour {
		return ErrRestoreWindow
	}
	return restore(p.db.WithContext(p.ctx), id)
}

func (p *Processor) PendingCount() (int64, error) {
	return countByStatus(p.db.WithContext(p.ctx), "pending")
}

func (p *Processor) CompletedTodayCount() (int64, error) {
	return countCompletedToday(p.db.WithContext(p.ctx))
}

func (p *Processor) OverdueCount() (int64, error) {
	return countOverdue(p.db.WithContext(p.ctx))
}
