package entry

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotScheduled = errors.New("date is not on the item's schedule; can only skip scheduled days")
	ErrItemNotFound = errors.New("tracking item not found")
)

// WithScheduled pairs an entry model with whether the entry's date falls on
// the tracking item's effective schedule. The "scheduled" flag is a derived
// projection used by the REST layer; it is not part of the persisted Entry.
type WithScheduled struct {
	Entry     Model
	Scheduled bool
}

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// CreateOrUpdate upserts a tracking entry for the given item and date. It
// resolves the tracking item internally so the handler does not have to reach
// across domain boundaries to fetch ScaleType / ScaleConfig before validating
// the submitted value.
//
// Returns the resulting entry, a "created" flag (true on insert, false on
// update), and a "scheduled" flag computed from the item's effective schedule
// snapshot for the entry date.
func (p *Processor) CreateOrUpdate(tenantID, userID, itemID uuid.UUID, dateStr string, value json.RawMessage, note *string) (Model, bool, bool, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return Model{}, false, false, ErrDateRequired
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if date.After(today) {
		return Model{}, false, false, ErrFutureDate
	}

	item, err := trackingitem.GetByID(itemID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, false, false, ErrItemNotFound
	}

	if err := ValidateValue(item.ScaleType, value, item.ScaleConfig); err != nil {
		return Model{}, false, false, err
	}

	if note != nil && len(*note) > 500 {
		return Model{}, false, false, ErrNoteTooLong
	}

	scheduled := p.isScheduledForDate(itemID, date)

	existing, err := GetByItemAndDate(itemID, date)(p.db.WithContext(p.ctx))()
	if err == nil {
		existing.Value = value
		existing.Skipped = false
		existing.Note = note
		if err := updateEntry(p.db.WithContext(p.ctx), &existing); err != nil {
			return Model{}, false, false, err
		}
		m, err := Make(existing)
		return m, false, scheduled, err
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
		return Model{}, false, false, err
	}
	m, err := Make(e)
	return m, true, scheduled, err
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

// Skip marks the entry for the given item/date as skipped. The processor
// verifies the tracking item exists and that the date falls on the item's
// effective schedule — only scheduled days may be skipped.
func (p *Processor) Skip(tenantID, userID, itemID uuid.UUID, dateStr string) (Model, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return Model{}, ErrDateRequired
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if date.After(today) {
		return Model{}, ErrFutureDate
	}

	if _, err := trackingitem.GetByID(itemID)(p.db.WithContext(p.ctx))(); err != nil {
		return Model{}, ErrItemNotFound
	}

	if !p.isScheduledForDate(itemID, date) {
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

// ListByMonthWithScheduled returns the entries for the given user/month paired
// with the "scheduled" projection (whether each entry's date falls on its
// item's effective schedule). The schedule lookup is cached per item to avoid
// re-querying within a single response.
func (p *Processor) ListByMonthWithScheduled(userID uuid.UUID, monthStr string) ([]WithScheduled, error) {
	models, err := p.ListByMonth(userID, monthStr)
	if err != nil {
		return nil, err
	}

	scheduledByItem := make(map[uuid.UUID][]int)
	results := make([]WithScheduled, len(models))
	for i, m := range models {
		sched, ok := scheduledByItem[m.TrackingItemID()]
		if !ok {
			sched = p.scheduleForItem(m.TrackingItemID(), m.Date())
			scheduledByItem[m.TrackingItemID()] = sched
		}
		results[i] = WithScheduled{Entry: m, Scheduled: matchesSchedule(sched, m.Date())}
	}
	return results, nil
}

// isScheduledForDate reports whether the given date falls on the tracking
// item's effective schedule snapshot. A nil/empty schedule means "every day".
// On any lookup error we conservatively treat the day as scheduled so a
// transient read failure does not block writes.
func (p *Processor) isScheduledForDate(itemID uuid.UUID, date time.Time) bool {
	sched := p.scheduleForItem(itemID, date)
	return matchesSchedule(sched, date)
}

func (p *Processor) scheduleForItem(itemID uuid.UUID, date time.Time) []int {
	snap, err := schedule.GetEffectiveSchedule(itemID, date)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil
	}
	m, err := schedule.Make(snap)
	if err != nil {
		return nil
	}
	return m.Schedule()
}

func matchesSchedule(sched []int, date time.Time) bool {
	if len(sched) == 0 {
		return true
	}
	dow := int(date.Weekday())
	for _, d := range sched {
		if d == dow {
			return true
		}
	}
	return false
}

func parseMonth(monthStr string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("month must be in YYYY-MM format")
	}
	end := start.AddDate(0, 1, -1)
	return start, end, nil
}
