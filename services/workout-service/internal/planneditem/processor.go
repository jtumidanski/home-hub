package planneditem

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound          = errors.New("planned item not found")
	ErrExerciseNotFound  = errors.New("exercise not found")
	ErrExerciseDeleted   = errors.New("cannot plan a soft-deleted exercise")
	ErrExerciseMismatch  = errors.New("exercise does not belong to this user")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// AddInput captures the per-item attributes for both the single-add and
// bulk-add endpoints. `Position` is a pointer so callers can omit it (the
// processor assigns the next available slot for the day).
type AddInput struct {
	ExerciseID             uuid.UUID
	DayOfWeek              int
	Position               *int
	PlannedSets            *int
	PlannedReps            *int
	PlannedWeight          *float64
	PlannedWeightUnit      *string
	PlannedDurationSeconds *int
	PlannedDistance        *float64
	PlannedDistanceUnit    *string
	Notes                  *string
}

// Add inserts a single planned item into the supplied week. The caller must
// have already lazily created the week row. The default planned values are
// seeded from the exercise's defaults when the input omits them.
func (p *Processor) Add(tenantID, userID, weekID uuid.UUID, in AddInput) (Model, error) {
	return p.addWithTx(p.db.WithContext(p.ctx), tenantID, userID, weekID, in)
}

// BulkAdd inserts every supplied item in a single transaction. Validation runs
// per-item but the whole batch is atomic — any failure rolls everything back.
func (p *Processor) BulkAdd(tenantID, userID, weekID uuid.UUID, items []AddInput) ([]Model, error) {
	var out []Model
	err := p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		for _, in := range items {
			m, err := p.addWithTx(tx, tenantID, userID, weekID, in)
			if err != nil {
				return err
			}
			out = append(out, m)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (p *Processor) addWithTx(tx *gorm.DB, tenantID, userID, weekID uuid.UUID, in AddInput) (Model, error) {
	ex, err := exercise.GetByID(in.ExerciseID)(tx)()
	if err != nil {
		// Either the exercise doesn't exist or it's soft-deleted. Distinguish
		// the two so the REST layer can return 404 vs 422 per §4.3.
		if _, errDel := exercise.GetByIDIncludeDeleted(in.ExerciseID)(tx)(); errDel == nil {
			return Model{}, ErrExerciseDeleted
		}
		return Model{}, ErrExerciseNotFound
	}
	if ex.UserId != userID {
		return Model{}, ErrExerciseMismatch
	}

	position := 0
	if in.Position != nil {
		position = *in.Position
	} else {
		max, err := MaxPositionForDay(tx, weekID, in.DayOfWeek)
		if err != nil {
			return Model{}, err
		}
		position = max + 1
	}

	// Apply exercise defaults whenever the request omitted a planned field.
	in = applyExerciseDefaults(in, ex)

	if _, err := NewBuilder().
		SetTenantID(tenantID).
		SetUserID(userID).
		SetWeekID(weekID).
		SetExerciseID(in.ExerciseID).
		SetDayOfWeek(in.DayOfWeek).
		SetPosition(position).
		SetPlannedSets(in.PlannedSets).
		SetPlannedReps(in.PlannedReps).
		SetPlannedWeight(in.PlannedWeight).
		SetPlannedWeightUnit(in.PlannedWeightUnit).
		SetPlannedDurationSeconds(in.PlannedDurationSeconds).
		SetPlannedDistance(in.PlannedDistance).
		SetPlannedDistanceUnit(in.PlannedDistanceUnit).
		SetNotes(in.Notes).
		Build(); err != nil {
		return Model{}, err
	}

	e := Entity{
		TenantId:               tenantID,
		UserId:                 userID,
		WeekId:                 weekID,
		ExerciseId:             in.ExerciseID,
		DayOfWeek:              in.DayOfWeek,
		Position:               position,
		PlannedSets:            in.PlannedSets,
		PlannedReps:            in.PlannedReps,
		PlannedWeight:          in.PlannedWeight,
		PlannedWeightUnit:      in.PlannedWeightUnit,
		PlannedDurationSeconds: in.PlannedDurationSeconds,
		PlannedDistance:        in.PlannedDistance,
		PlannedDistanceUnit:    in.PlannedDistanceUnit,
		Notes:                  in.Notes,
	}
	if err := createPlannedItem(tx, &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func applyExerciseDefaults(in AddInput, ex exercise.Entity) AddInput {
	if in.PlannedSets == nil {
		in.PlannedSets = ex.DefaultSets
	}
	if in.PlannedReps == nil {
		in.PlannedReps = ex.DefaultReps
	}
	if in.PlannedWeight == nil {
		in.PlannedWeight = ex.DefaultWeight
	}
	if in.PlannedWeightUnit == nil {
		in.PlannedWeightUnit = ex.DefaultWeightUnit
	}
	if in.PlannedDurationSeconds == nil {
		in.PlannedDurationSeconds = ex.DefaultDurationSeconds
	}
	if in.PlannedDistance == nil {
		in.PlannedDistance = ex.DefaultDistance
	}
	if in.PlannedDistanceUnit == nil {
		in.PlannedDistanceUnit = ex.DefaultDistanceUnit
	}
	return in
}

// UpdateInput is the partial update payload from PATCH .../items/{itemId}.
// All fields are pointers so the processor can distinguish "omitted" from
// "explicit null" — explicit null clears the planned value.
type UpdateInput struct {
	DayOfWeek              *int
	Position               *int
	PlannedSets            *int
	PlannedReps            *int
	PlannedWeight          *float64
	PlannedWeightUnit      *string
	PlannedDurationSeconds *int
	PlannedDistance        *float64
	PlannedDistanceUnit    *string
	Notes                  *string
}

func (p *Processor) Update(id uuid.UUID, in UpdateInput) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if in.DayOfWeek != nil {
		e.DayOfWeek = *in.DayOfWeek
	}
	if in.Position != nil {
		e.Position = *in.Position
	}
	if in.PlannedSets != nil {
		e.PlannedSets = in.PlannedSets
	}
	if in.PlannedReps != nil {
		e.PlannedReps = in.PlannedReps
	}
	if in.PlannedWeight != nil {
		e.PlannedWeight = in.PlannedWeight
	}
	if in.PlannedWeightUnit != nil {
		e.PlannedWeightUnit = in.PlannedWeightUnit
	}
	if in.PlannedDurationSeconds != nil {
		e.PlannedDurationSeconds = in.PlannedDurationSeconds
	}
	if in.PlannedDistance != nil {
		e.PlannedDistance = in.PlannedDistance
	}
	if in.PlannedDistanceUnit != nil {
		e.PlannedDistanceUnit = in.PlannedDistanceUnit
	}
	if in.Notes != nil {
		e.Notes = in.Notes
	}

	if _, err := Make(e); err != nil {
		return Model{}, err
	}

	if err := updatePlannedItem(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	if _, err := GetByID(id)(p.db.WithContext(p.ctx))(); err != nil {
		return ErrNotFound
	}
	return deletePlannedItem(p.db.WithContext(p.ctx), id)
}

// ReorderEntry is one row in the reorder request payload.
type ReorderEntry struct {
	ItemID    uuid.UUID
	DayOfWeek int
	Position  int
}

// Reorder applies a list of (itemId → day, position) updates atomically. Used
// by the drag-and-drop weekly planner. The new updated_at timestamp is
// computed in Go so the same query works under both Postgres and the
// sqlite-backed test harness.
func (p *Processor) Reorder(weekID uuid.UUID, entries []ReorderEntry) error {
	return p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		for _, r := range entries {
			if r.DayOfWeek < 0 || r.DayOfWeek > 6 {
				return ErrInvalidDayOfWeek
			}
			if r.Position < 0 {
				return ErrInvalidPosition
			}
			res := tx.Model(&Entity{}).
				Where("id = ? AND week_id = ?", r.ItemID, weekID).
				Updates(map[string]any{
					"day_of_week": r.DayOfWeek,
					"position":    r.Position,
					"updated_at":  now,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ErrNotFound
			}
		}
		return nil
	})
}
