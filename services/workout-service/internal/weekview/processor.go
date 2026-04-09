// Package weekview's Processor owns the cross-domain orchestration that
// produces the embedded "week with items" projection and executes the
// copy-from-previous workflow. Domain processors (week, planneditem,
// performance) remain the source of truth for their own state; this
// processor only stitches them together for the composite endpoints.
package weekview

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Domain-level errors surfaced by this package. The handlers map them onto
// the §4.2 status codes (404 / 409 / 400 / 422).
var (
	ErrCopyTargetNotEmpty = errors.New("target week already has planned items")
	ErrCopyNoSource       = errors.New("no prior week with planned items found")
	ErrInvalidCopyMode    = errors.New("mode must be 'planned' or 'actual'")
)

const (
	CopyModePlanned = "planned"
	CopyModeActual  = "actual"
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// LoadWeekDocument resolves the week + its embedded items into a JSON:API
// document. Used by every read and post-mutation render path.
func (p *Processor) LoadWeekDocument(weekModel week.Model) (Document, error) {
	items, err := AssembleItems(p.db.WithContext(p.ctx), weekModel.Id())
	if err != nil {
		return Document{}, err
	}
	return BuildDocument(weekModel, items), nil
}

// Copy implements `POST /workouts/weeks/{weekStart}/copy`. The week is
// lazily created if missing; the source is the most recent prior week with
// planned items.
//
//   - planned: clone every planned item from the source week verbatim.
//   - actual:  clone the structure but seed each planned_* field from the
//     corresponding performance row's actuals when present, falling back to
//     the source planned values otherwise. Per-set rows are collapsed using
//     the §5.3 rule (count, max-reps, max-weight).
func (p *Processor) Copy(tenantID, userID uuid.UUID, weekStart time.Time, mode string) (week.Model, error) {
	if mode != CopyModePlanned && mode != CopyModeActual {
		return week.Model{}, ErrInvalidCopyMode
	}

	weekProc := week.NewProcessor(p.l, p.ctx, p.db)
	target, err := weekProc.EnsureExists(tenantID, userID, weekStart)
	if err != nil {
		return week.Model{}, err
	}

	// Reject if the target already has any planned items.
	existing, err := planneditem.GetByWeek(target.Id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return week.Model{}, err
	}
	if len(existing) > 0 {
		return week.Model{}, ErrCopyTargetNotEmpty
	}

	source, err := week.GetMostRecentPriorWithItems(p.db.WithContext(p.ctx), userID, target.WeekStartDate)
	if err != nil {
		return week.Model{}, ErrCopyNoSource
	}
	sourceItems, err := planneditem.GetByWeek(source.Id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return week.Model{}, err
	}
	if len(sourceItems) == 0 {
		return week.Model{}, ErrCopyNoSource
	}

	itemIDs := make([]uuid.UUID, 0, len(sourceItems))
	for _, it := range sourceItems {
		itemIDs = append(itemIDs, it.Id)
	}
	perfMap, setMap, err := performance.LoadByPlannedItems(p.db.WithContext(p.ctx), itemIDs)
	if err != nil {
		return week.Model{}, err
	}

	// Insert clones inside a single transaction so a partial copy can never
	// leak past a mid-loop failure.
	err = p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		for _, src := range sourceItems {
			cloned := src
			cloned.Id = uuid.Nil // re-issue
			cloned.WeekId = target.Id

			if mode == CopyModeActual {
				if perf, ok := perfMap[src.Id]; ok {
					applyActualsAsPlanned(&cloned, perf, setMap[perf.Id])
				}
			}

			ent := planneditem.Entity{
				TenantId:               tenantID,
				UserId:                 userID,
				WeekId:                 target.Id,
				ExerciseId:             cloned.ExerciseId,
				DayOfWeek:              cloned.DayOfWeek,
				Position:               cloned.Position,
				PlannedSets:            cloned.PlannedSets,
				PlannedReps:            cloned.PlannedReps,
				PlannedWeight:          cloned.PlannedWeight,
				PlannedWeightUnit:      cloned.PlannedWeightUnit,
				PlannedDurationSeconds: cloned.PlannedDurationSeconds,
				PlannedDistance:        cloned.PlannedDistance,
				PlannedDistanceUnit:    cloned.PlannedDistanceUnit,
				Notes:                  cloned.Notes,
			}
			if err := tx.Create(&ent).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return week.Model{}, err
	}
	return week.Make(target)
}

// applyActualsAsPlanned mutates `cloned` so its planned_* fields reflect the
// source week's actuals. Per-set actuals are collapsed using the §5.3 rule
// (count → planned_sets, max reps → planned_reps, max weight → planned_weight).
// Fields with no actual fall back to the source planned value, which `cloned`
// already carries from the verbatim clone.
func applyActualsAsPlanned(cloned *planneditem.Entity, perf performance.Entity, sets []performance.SetEntity) {
	if perf.Mode == performance.ModePerSet && len(sets) > 0 {
		count := len(sets)
		maxReps := 0
		var maxWeight float64
		for _, s := range sets {
			if s.Reps > maxReps {
				maxReps = s.Reps
			}
			if s.Weight > maxWeight {
				maxWeight = s.Weight
			}
		}
		cs := count
		mr := maxReps
		mw := maxWeight
		cloned.PlannedSets = &cs
		cloned.PlannedReps = &mr
		cloned.PlannedWeight = &mw
		if perf.WeightUnit != nil {
			cloned.PlannedWeightUnit = perf.WeightUnit
		}
		return
	}
	// Summary mode: copy whichever actuals are present.
	if perf.ActualSets != nil {
		cloned.PlannedSets = perf.ActualSets
	}
	if perf.ActualReps != nil {
		cloned.PlannedReps = perf.ActualReps
	}
	if perf.ActualWeight != nil {
		cloned.PlannedWeight = perf.ActualWeight
	}
	if perf.WeightUnit != nil {
		cloned.PlannedWeightUnit = perf.WeightUnit
	}
	if perf.ActualDurationSeconds != nil {
		cloned.PlannedDurationSeconds = perf.ActualDurationSeconds
	}
	if perf.ActualDistance != nil {
		cloned.PlannedDistance = perf.ActualDistance
	}
	if perf.ActualDistanceUnit != nil {
		cloned.PlannedDistanceUnit = perf.ActualDistanceUnit
	}
}
