package performance

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrPlannedItemNotFound  = errors.New("planned item not found")
	ErrPerSetNotAllowed     = errors.New("per-set logging is only valid for strength items")
	ErrSummaryWhilePerSet   = errors.New("cannot write summary actuals while per-set rows exist; collapse first")
	ErrUnitChangeWithSets   = errors.New("cannot change weightUnit while per-set rows exist")
	ErrInvalidSetNumeric    = errors.New("set reps and weight must be non-negative")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// PatchInput is the summary-mode update payload from PATCH .../performance.
// Pointer fields preserve "omitted vs explicit null" so the processor can
// either leave a field unchanged or clear it.
type PatchInput struct {
	Status                *string
	WeightUnit            *string
	ActualSets            *int
	ActualReps            *int
	ActualWeight          *float64
	ActualDurationSeconds *int
	ActualDistance        *float64
	ActualDistanceUnit    *string
	Notes                 *string
}

// hasActuals reports whether any actuals fields were sent on the patch. The
// state-derivation rule in api-contracts §5.1 only fires when actuals were
// supplied without an explicit status.
func (in PatchInput) hasActuals() bool {
	return in.ActualSets != nil || in.ActualReps != nil || in.ActualWeight != nil ||
		in.ActualDurationSeconds != nil || in.ActualDistance != nil
}

// Patch applies a summary-mode update. Implements the state machine in PRD
// §4.4.1 plus the reject paths defined in api-contracts §5.1.
func (p *Processor) Patch(tenantID, userID uuid.UUID, plannedItemID uuid.UUID, in PatchInput) (Model, []SetModel, error) {
	pi, err := planneditem.GetByID(plannedItemID)(p.db.WithContext(p.ctx))()
	if err != nil || pi.UserId != userID {
		return Model{}, nil, ErrPlannedItemNotFound
	}

	e, perfExists := p.findOrEmpty(tenantID, userID, plannedItemID)

	// Per-set guardrails: explicit summary writes are forbidden in per_set
	// mode and weight-unit changes are forbidden whenever per-set rows exist.
	if perfExists && e.Mode == ModePerSet {
		if in.hasActuals() {
			return Model{}, nil, ErrSummaryWhilePerSet
		}
		if in.WeightUnit != nil {
			rows, err := loadSets(p.db.WithContext(p.ctx), e.Id)
			if err != nil {
				return Model{}, nil, err
			}
			if len(rows) > 0 {
				return Model{}, nil, ErrUnitChangeWithSets
			}
		}
	}

	// Apply field changes.
	if in.WeightUnit != nil {
		e.WeightUnit = in.WeightUnit
	}
	if in.Notes != nil {
		e.Notes = in.Notes
	}
	if in.hasActuals() {
		if in.ActualSets != nil {
			e.ActualSets = in.ActualSets
		}
		if in.ActualReps != nil {
			e.ActualReps = in.ActualReps
		}
		if in.ActualWeight != nil {
			e.ActualWeight = in.ActualWeight
		}
		if in.ActualDurationSeconds != nil {
			e.ActualDurationSeconds = in.ActualDurationSeconds
		}
		if in.ActualDistance != nil {
			e.ActualDistance = in.ActualDistance
		}
		if in.ActualDistanceUnit != nil {
			e.ActualDistanceUnit = in.ActualDistanceUnit
		}
	}

	// State derivation per §5.1.
	prev := e.Status
	if in.Status != nil {
		e.Status = applyExplicitStatus(prev, *in.Status, hasAnyActuals(e))
	} else if in.hasActuals() {
		e.Status = deriveStatusFromActuals(prev)
	}

	// Validate the merged state through the builder so unit/numeric/status
	// invariants are enforced even when the patch arrived in slices.
	m, err := Make(e)
	if err != nil {
		return Model{}, nil, err
	}

	if !perfExists {
		if err := createPerformance(p.db.WithContext(p.ctx), &e); err != nil {
			return Model{}, nil, err
		}
	} else {
		if err := updatePerformance(p.db.WithContext(p.ctx), &e); err != nil {
			return Model{}, nil, err
		}
	}
	m, err = Make(e)
	if err != nil {
		return Model{}, nil, err
	}
	return m, nil, nil
}

// applyExplicitStatus implements the explicit-transition rules in §5.1.
func applyExplicitStatus(prev, requested string, hasActuals bool) string {
	switch requested {
	case StatusDone:
		return StatusDone
	case StatusSkipped:
		return StatusSkipped
	case StatusPending:
		// `pending` from `done` becomes `partial` if there are actuals,
		// `pending` from `skipped` becomes `pending` (unskip).
		switch prev {
		case StatusDone:
			if hasActuals {
				return StatusPartial
			}
			return StatusPending
		case StatusSkipped:
			return StatusPending
		default:
			return StatusPending
		}
	case StatusPartial:
		return StatusPartial
	}
	return prev
}

// deriveStatusFromActuals implements the auto-derivation rules in §5.1 for
// the case where the client sent actuals without specifying a status.
func deriveStatusFromActuals(prev string) string {
	switch prev {
	case StatusPending, StatusSkipped:
		return StatusPartial
	case StatusPartial:
		return StatusPartial
	case StatusDone:
		return StatusDone
	}
	return StatusPartial
}

func hasAnyActuals(e Entity) bool {
	return e.ActualSets != nil || e.ActualReps != nil || e.ActualWeight != nil ||
		e.ActualDurationSeconds != nil || e.ActualDistance != nil
}

func (p *Processor) findOrEmpty(tenantID, userID, plannedItemID uuid.UUID) (Entity, bool) {
	e, err := GetByPlannedItem(plannedItemID)(p.db.WithContext(p.ctx))()
	if err == nil {
		return e, true
	}
	return Entity{
		TenantId:      tenantID,
		UserId:        userID,
		PlannedItemId: plannedItemID,
		Status:        StatusPending,
		Mode:          ModeSummary,
	}, false
}

// SetInput is one row in the per-set replace payload. The server assigns
// `set_number` from the array index.
type SetInput struct {
	Reps   int
	Weight float64
}

// ReplaceSets replaces all per-set rows for a performance. Switches the
// performance into per_set mode, clears summary actuals, and is rejected for
// non-strength items per §5.2.
func (p *Processor) ReplaceSets(tenantID, userID uuid.UUID, plannedItemID uuid.UUID, weightUnit string, sets []SetInput) (Model, []SetModel, error) {
	pi, err := planneditem.GetByID(plannedItemID)(p.db.WithContext(p.ctx))()
	if err != nil || pi.UserId != userID {
		return Model{}, nil, ErrPlannedItemNotFound
	}
	ex, err := exercise.GetByIDIncludeDeleted(pi.ExerciseId)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, nil, ErrPlannedItemNotFound
	}
	if ex.Kind != exercise.KindStrength {
		return Model{}, nil, ErrPerSetNotAllowed
	}
	if !ValidWeightUnit(weightUnit) {
		return Model{}, nil, ErrInvalidWeightUnit
	}
	for _, s := range sets {
		if s.Reps < 0 || s.Weight < 0 {
			return Model{}, nil, ErrInvalidSetNumeric
		}
	}

	e, exists := p.findOrEmpty(tenantID, userID, plannedItemID)
	wu := weightUnit
	e.WeightUnit = &wu
	e.Mode = ModePerSet
	// Clear summary actuals — per-set is the source of truth now.
	e.ActualSets = nil
	e.ActualReps = nil
	e.ActualWeight = nil

	// Promote pending → partial when sets land. The user explicitly logged
	// reps; the §4.4.1 transitions treat that as "log actuals → partial".
	if e.Status == StatusPending || e.Status == StatusSkipped {
		e.Status = StatusPartial
	}

	out := make([]SetModel, 0, len(sets))
	err = p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		if !exists {
			if err := createPerformance(tx, &e); err != nil {
				return err
			}
		} else {
			if err := updatePerformance(tx, &e); err != nil {
				return err
			}
			if err := deleteSetsForPerformance(tx, e.Id); err != nil {
				return err
			}
		}
		for i, s := range sets {
			row := SetEntity{
				TenantId:      tenantID,
				UserId:        userID,
				PerformanceId: e.Id,
				SetNumber:     i + 1,
				Reps:          s.Reps,
				Weight:        s.Weight,
			}
			if err := createSet(tx, &row); err != nil {
				return err
			}
			out = append(out, MakeSet(row))
		}
		return nil
	})
	if err != nil {
		return Model{}, nil, err
	}
	m, err := Make(e)
	if err != nil {
		return Model{}, nil, err
	}
	return m, out, nil
}

// CollapseSets switches the performance back to summary mode using the §5.3
// collapse rule (count, max-reps, max-weight). Returns 404 when the planned
// item or its performance is missing.
func (p *Processor) CollapseSets(userID uuid.UUID, plannedItemID uuid.UUID) (Model, error) {
	pi, err := planneditem.GetByID(plannedItemID)(p.db.WithContext(p.ctx))()
	if err != nil || pi.UserId != userID {
		return Model{}, ErrPlannedItemNotFound
	}
	e, err := GetByPlannedItem(plannedItemID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrPlannedItemNotFound
	}
	rows, err := loadSets(p.db.WithContext(p.ctx), e.Id)
	if err != nil {
		return Model{}, err
	}

	count := len(rows)
	maxReps := 0
	var maxWeight float64
	for _, r := range rows {
		if r.Reps > maxReps {
			maxReps = r.Reps
		}
		if r.Weight > maxWeight {
			maxWeight = r.Weight
		}
	}

	cs := count
	mr := maxReps
	mw := maxWeight
	e.Mode = ModeSummary
	e.ActualSets = &cs
	e.ActualReps = &mr
	e.ActualWeight = &mw

	err = p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		if err := deleteSetsForPerformance(tx, e.Id); err != nil {
			return err
		}
		return updatePerformance(tx, &e)
	})
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
