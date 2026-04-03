package entry

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotScheduled = errors.New("date is not on the item's schedule; can only skip scheduled days")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) CreateOrUpdate(tenantID, userID, itemID uuid.UUID, dateStr string, value json.RawMessage, note *string, scaleType string, scaleConfig json.RawMessage) (Model, bool, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return Model{}, false, ErrDateRequired
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if date.After(today) {
		return Model{}, false, ErrFutureDate
	}

	if err := ValidateValue(scaleType, value, scaleConfig); err != nil {
		return Model{}, false, err
	}

	if note != nil && len(*note) > 500 {
		return Model{}, false, ErrNoteTooLong
	}

	existing, err := GetByItemAndDate(itemID, date)(p.db.WithContext(p.ctx))()
	if err == nil {
		existing.Value = value
		existing.Skipped = false
		existing.Note = note
		if err := updateEntry(p.db.WithContext(p.ctx), &existing); err != nil {
			return Model{}, false, err
		}
		m, err := Make(existing)
		return m, false, err
	}

	e := Entity{
		TenantId:       tenantID,
		UserId:         userID,
		TrackingItemId: itemID,
		Date:           date,
		Value:          value,
		Skipped:        false,
		Note:           note,
	}
	if err := createEntry(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, false, err
	}
	m, err := Make(e)
	return m, true, err
}

func (p *Processor) Delete(itemID uuid.UUID, dateStr string) error {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ErrDateRequired
	}

	existing, err := GetByItemAndDate(itemID, date)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil
	}
	return deleteEntry(p.db.WithContext(p.ctx), existing.Id)
}

func (p *Processor) Skip(tenantID, userID, itemID uuid.UUID, dateStr string, isScheduled bool) (Model, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return Model{}, ErrDateRequired
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if date.After(today) {
		return Model{}, ErrFutureDate
	}

	if !isScheduled {
		return Model{}, ErrNotScheduled
	}

	existing, err := GetByItemAndDate(itemID, date)(p.db.WithContext(p.ctx))()
	if err == nil {
		existing.Value = nil
		existing.Skipped = true
		existing.Note = nil
		if err := updateEntry(p.db.WithContext(p.ctx), &existing); err != nil {
			return Model{}, err
		}
		return Make(existing)
	}

	e := Entity{
		TenantId:       tenantID,
		UserId:         userID,
		TrackingItemId: itemID,
		Date:           date,
		Value:          nil,
		Skipped:        true,
	}
	if err := createEntry(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) RemoveSkip(itemID uuid.UUID, dateStr string) error {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ErrDateRequired
	}

	existing, err := GetByItemAndDate(itemID, date)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil
	}
	if existing.Skipped {
		return deleteEntry(p.db.WithContext(p.ctx), existing.Id)
	}
	return nil
}

func (p *Processor) ListByMonth(userID uuid.UUID, monthStr string) ([]Model, error) {
	start, end, err := parseMonth(monthStr)
	if err != nil {
		return nil, err
	}

	entities, err := GetByUserAndMonth(userID, start, end)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}

	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		models[i] = m
	}
	return models, nil
}

func parseMonth(monthStr string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("month must be in YYYY-MM format")
	}
	end := start.AddDate(0, 1, -1)
	return start, end, nil
}
