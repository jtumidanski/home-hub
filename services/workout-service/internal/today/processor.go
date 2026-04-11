// Package today's Processor builds the per-day projection used by the
// `GET /workouts/today` endpoint. "Now" is resolved in the caller's
// timezone (see internal/tz); callers must pass a time.Time whose Location
// is already the resolved zone.
package today

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/weekview"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// Result is the projected today document. Empty `Items` is the legitimate
// "no plan" state — callers must NOT translate it into a 404.
type Result struct {
	Date          time.Time
	WeekStartDate time.Time
	DayOfWeek     int
	IsRestDay     bool
	Items         []weekview.ItemRest
}

// Today computes today's projection for a user. Returns an empty `Items`
// slice when the user has no week row for the current ISO week, never an
// error in that case.
func (p *Processor) Today(userID uuid.UUID, now time.Time) (Result, error) {
	// Compute the calendar day in the caller's zone, then anchor to UTC for
	// DB lookups so the DATE column ("type:date") compares as a plain date
	// irrespective of tz offset.
	dayOfWeek := (int(now.Weekday()) + 6) % 7 // Mon = 0
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekStart := today.AddDate(0, 0, -dayOfWeek)
	now = today

	weekProc := week.NewProcessor(p.l, p.ctx, p.db)
	m, err := weekProc.Get(userID, weekStart)
	if err != nil {
		if errors.Is(err, week.ErrNotFound) {
			return Result{
				Date:          now,
				WeekStartDate: weekStart,
				DayOfWeek:     dayOfWeek,
				IsRestDay:     false,
				Items:         []weekview.ItemRest{},
			}, nil
		}
		return Result{}, err
	}

	items, err := weekview.AssembleItems(p.db.WithContext(p.ctx), m.Id())
	if err != nil {
		return Result{}, err
	}

	// Filter to today's day-of-week.
	filtered := make([]weekview.ItemRest, 0, len(items))
	for _, it := range items {
		if it.DayOfWeek == dayOfWeek {
			filtered = append(filtered, it)
		}
	}

	isRestDay := false
	for _, d := range m.RestDayFlags() {
		if d == dayOfWeek {
			isRestDay = true
			break
		}
	}

	return Result{
		Date:          now,
		WeekStartDate: weekStart,
		DayOfWeek:     dayOfWeek,
		IsRestDay:     isRestDay,
		Items:         filtered,
	}, nil
}
