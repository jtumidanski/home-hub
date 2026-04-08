package schedule

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ErrNotFound is returned when no schedule snapshot exists for the requested
// (item, date) tuple.
var ErrNotFound = errors.New("schedule snapshot not found")

// Processor is the orchestration entry point for schedule reads and writes.
// Other domain processors must invoke these methods rather than calling the
// package-level providers/administrator helpers directly so the schedule
// package owns its layering rules.
type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// GetEffective returns the schedule snapshot in effect for the given item on
// the given date. Returns ErrNotFound if no snapshot covers the date.
func (p *Processor) GetEffective(itemID uuid.UUID, date time.Time) (Model, error) {
	e, err := GetEffectiveSchedule(itemID, date)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return Make(e)
}

// GetHistory returns every schedule snapshot for the given item, ordered by
// effective date ascending.
func (p *Processor) GetHistory(itemID uuid.UUID) ([]Model, error) {
	entities, err := GetByTrackingItemID(itemID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	models := make([]Model, 0, len(entities))
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

// GetHistoriesByItems returns the schedule snapshots for the given item IDs
// keyed by item ID. Used by the month report which renders multiple items at
// once.
func (p *Processor) GetHistoriesByItems(itemIDs []uuid.UUID) (map[uuid.UUID][]Model, error) {
	if len(itemIDs) == 0 {
		return map[uuid.UUID][]Model{}, nil
	}
	entities, err := GetByTrackingItemIDs(itemIDs)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID][]Model)
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			continue
		}
		result[m.TrackingItemID()] = append(result[m.TrackingItemID()], m)
	}
	return result, nil
}

// CreateSnapshot upserts a schedule snapshot for the given item / effective
// date. Construct the processor with a transaction-scoped *gorm.DB when
// calling from inside a Tx.
func (p *Processor) CreateSnapshot(itemID uuid.UUID, days []int, effectiveDate time.Time) (Model, error) {
	for _, d := range days {
		if d < 0 || d > 6 {
			return Model{}, ErrInvalidScheduleDay
		}
	}
	e, err := CreateSnapshot(p.db.WithContext(p.ctx), itemID, days, effectiveDate)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
