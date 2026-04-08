package today

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Result is the orchestrated payload for the "today" view: the user's tracking
// items that are scheduled to occur on `Date`, paired with any entries already
// logged for that date. The REST layer transforms this into a JSON:API
// document; the processor stays free of HTTP concerns.
type Result struct {
	Date    time.Time
	Items   []trackingitem.Model
	Entries []entry.Model
}

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// Today returns the tracking items scheduled for `date` for the given user
// along with any entries already recorded for that date. Cross-domain reads
// (trackingitem, schedule, entry) are kept inside the processor so handlers
// never reach across packages for orchestration.
func (p *Processor) Today(userID uuid.UUID, date time.Time) (Result, error) {
	day := date.UTC().Truncate(24 * time.Hour)
	dow := int(day.Weekday())

	itemEntities, err := trackingitem.GetAllByUser(userID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Result{}, err
	}

	var scheduledItems []trackingitem.Model
	for _, e := range itemEntities {
		m, err := trackingitem.Make(e)
		if err != nil {
			p.l.WithError(err).WithField("item_id", e.Id).Warn("Skipping unreadable tracking item")
			continue
		}

		snap, err := schedule.GetEffectiveSchedule(m.Id(), day)(p.db.WithContext(p.ctx))()
		if err != nil {
			// No effective snapshot — item has not started or schedule is missing.
			continue
		}
		sm, err := schedule.Make(snap)
		if err != nil {
			p.l.WithError(err).WithField("item_id", m.Id()).Warn("Skipping unreadable schedule snapshot")
			continue
		}
		if !scheduleMatches(sm.Schedule(), dow) {
			continue
		}
		scheduledItems = append(scheduledItems, m)
	}

	var todayEntries []entry.Model
	for _, item := range scheduledItems {
		e, err := entry.GetByItemAndDate(item.Id(), day)(p.db.WithContext(p.ctx))()
		if err != nil {
			continue
		}
		em, err := entry.Make(e)
		if err != nil {
			p.l.WithError(err).WithField("item_id", item.Id()).Warn("Skipping unreadable today entry")
			continue
		}
		todayEntries = append(todayEntries, em)
	}

	return Result{Date: day, Items: scheduledItems, Entries: todayEntries}, nil
}

// scheduleMatches reports whether the given day-of-week is on the schedule.
// An empty schedule is treated as "every day".
func scheduleMatches(sched []int, dow int) bool {
	if len(sched) == 0 {
		return true
	}
	for _, d := range sched {
		if d == dow {
			return true
		}
	}
	return false
}
