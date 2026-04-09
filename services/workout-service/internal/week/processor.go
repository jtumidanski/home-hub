package week

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("week not found")

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) Get(userID uuid.UUID, weekStart time.Time) (Model, error) {
	weekStart = NormalizeToMonday(weekStart)
	e, err := GetByUserAndStart(userID, weekStart)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return Make(e)
}

// EnsureExists is the lazy-create primitive shared by every week mutation. It
// returns the existing row if present, or creates an empty one for the
// normalized Monday otherwise. The (tenant, user, week_start) unique index
// makes a parallel race safe — the second writer falls back to fetching the
// row that the first writer committed.
func (p *Processor) EnsureExists(tenantID, userID uuid.UUID, weekStart time.Time) (Entity, error) {
	weekStart = NormalizeToMonday(weekStart)
	if e, err := GetByUserAndStart(userID, weekStart)(p.db.WithContext(p.ctx))(); err == nil {
		return e, nil
	}
	e := Entity{
		TenantId:      tenantID,
		UserId:        userID,
		WeekStartDate: weekStart,
		RestDayFlags:  json.RawMessage("[]"),
	}
	if err := createWeek(p.db.WithContext(p.ctx), &e); err != nil {
		// Race-recovery: another writer just created it.
		if existing, err2 := GetByUserAndStart(userID, weekStart)(p.db.WithContext(p.ctx))(); err2 == nil {
			return existing, nil
		}
		return Entity{}, err
	}
	return e, nil
}

// PatchRestDayFlags is the only patchable field on weeks today. The endpoint
// lazily creates the week row so the caller does not need to seed it first.
func (p *Processor) PatchRestDayFlags(tenantID, userID uuid.UUID, weekStart time.Time, flags []int) (Model, error) {
	if err := ValidateRestDayFlags(flags); err != nil {
		return Model{}, err
	}
	e, err := p.EnsureExists(tenantID, userID, weekStart)
	if err != nil {
		return Model{}, err
	}
	jb, err := json.Marshal(flags)
	if err != nil {
		return Model{}, err
	}
	e.RestDayFlags = jb
	if err := updateWeek(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

// DB exposes the underlying gorm handle so caller packages (planneditem,
// performance) that operate inside week-scoped routes can share the same
// transaction or read-through helpers without re-resolving the connection.
func (p *Processor) DB() *gorm.DB { return p.db.WithContext(p.ctx) }
