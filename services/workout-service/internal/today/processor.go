// Package today's Processor builds the per-day projection used by the
// `GET /workouts/today` endpoint. The current day is computed in UTC,
// matching `tracker-service/today`. (PRD §6 calls for the user's TZ; we keep
// parity with tracker-service rather than reinventing TZ resolution here.
// When account-service grows a shared TZ helper, both Today endpoints can
// adopt it together.)
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
	now = now.UTC().Truncate(24 * time.Hour)
	weekStart := week.NormalizeToMonday(now)
	dayOfWeek := (int(now.Weekday()) + 6) % 7 // Mon = 0

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
