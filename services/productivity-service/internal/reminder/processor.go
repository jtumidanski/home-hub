package reminder

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var allowedSnoozeDurations = map[int]bool{10: true, 30: true, 60: true}

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

func (p *Processor) Create(tenantID, householdID uuid.UUID, title, notes string, scheduledFor time.Time) (Model, error) {
	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, title, notes, scheduledFor)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(id uuid.UUID, title, notes string, scheduledFor time.Time) (Model, error) {
	e, err := update(p.db.WithContext(p.ctx), id, title, notes, scheduledFor)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	return deleteByID(p.db.WithContext(p.ctx), id)
}

func (p *Processor) Dismiss(id uuid.UUID) error {
	return dismiss(p.db.WithContext(p.ctx), id)
}

func (p *Processor) Snooze(id uuid.UUID, durationMinutes int) (time.Time, error) {
	if !allowedSnoozeDurations[durationMinutes] {
		return time.Time{}, errors.New("invalid snooze duration: must be 10, 30, or 60 minutes")
	}
	snoozedUntil := time.Now().UTC().Add(time.Duration(durationMinutes) * time.Minute)
	if err := snooze(p.db.WithContext(p.ctx), id, snoozedUntil); err != nil {
		return time.Time{}, err
	}
	return snoozedUntil, nil
}

func (p *Processor) DueNowCount() (int64, error) {
	return countDueNow(p.db.WithContext(p.ctx))
}

func (p *Processor) UpcomingCount() (int64, error) {
	return countUpcoming(p.db.WithContext(p.ctx))
}

func (p *Processor) SnoozedCount() (int64, error) {
	return countSnoozed(p.db.WithContext(p.ctx))
}
