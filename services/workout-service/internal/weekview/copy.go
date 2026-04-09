package weekview

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	errCopyTargetNotEmpty = errors.New("target week already has planned items")
	errCopyNoSource       = errors.New("no prior week with planned items found")
)

// CopyHandler handles POST /workouts/weeks/{weekStart}/copy. The handler is
// exposed as a separate function rather than registered inside
// InitializeRoutes so the file/route ownership stays in copy.go and a future
// reader sees it next to the copy logic.
func CopyHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			weekStart, err := parseWeekStart(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}
			body, _ := io.ReadAll(r.Body)
			var env struct {
				Data struct {
					Attributes struct {
						Mode string `json:"mode"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}
			mode := env.Data.Attributes.Mode
			if mode != "planned" && mode != "actual" {
				server.WriteError(w, http.StatusBadRequest, "Validation Failed", "mode must be 'planned' or 'actual'")
				return
			}

			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			created, err := copyWeek(r.Context(), d.Logger(), db, t.Id(), t.UserId(), weekStart, mode)
			if err != nil {
				switch {
				case errors.Is(err, errCopyTargetNotEmpty):
					server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
				case errors.Is(err, errCopyNoSource):
					server.WriteError(w, http.StatusNotFound, "Not Found", err.Error())
				default:
					d.Logger().WithError(err).Error("Failed to copy week")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
				}
				return
			}
			renderWeek(w, d.Logger(), weekProc, created, http.StatusCreated)
		}
	}
}

// copyWeek implements the planned|actual semantics in PRD §4.5 and
// api-contracts §4.2:
//
//   - planned: clone every planned item from the source week verbatim.
//   - actual:  clone the structure but seed each planned_* field from the
//     corresponding performance row's actuals when present, falling back to
//     the source planned values otherwise. Per-set rows are collapsed using
//     the §5.3 rule (count, max-reps, max-weight).
func copyWeek(ctx context.Context, l logrus.FieldLogger, db *gorm.DB, tenantID, userID uuid.UUID, weekStart time.Time, mode string) (week.Model, error) {
	weekProc := week.NewProcessor(l, ctx, db)
	target, err := weekProc.EnsureExists(tenantID, userID, weekStart)
	if err != nil {
		return week.Model{}, err
	}

	// Reject if the target already has any planned items.
	existing, err := planneditem.GetByWeek(target.Id)(db.WithContext(ctx))()
	if err != nil {
		return week.Model{}, err
	}
	if len(existing) > 0 {
		return week.Model{}, errCopyTargetNotEmpty
	}

	source, err := week.GetMostRecentPriorWithItems(db.WithContext(ctx), userID, target.WeekStartDate)
	if err != nil {
		return week.Model{}, errCopyNoSource
	}
	sourceItems, err := planneditem.GetByWeek(source.Id)(db.WithContext(ctx))()
	if err != nil {
		return week.Model{}, err
	}
	if len(sourceItems) == 0 {
		return week.Model{}, errCopyNoSource
	}

	itemIDs := make([]uuid.UUID, 0, len(sourceItems))
	for _, it := range sourceItems {
		itemIDs = append(itemIDs, it.Id)
	}
	perfMap, setMap, err := performance.LoadByPlannedItems(db.WithContext(ctx), itemIDs)
	if err != nil {
		return week.Model{}, err
	}

	// Insert clones inside a single transaction so a partial copy can never
	// leak past a mid-loop failure.
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, src := range sourceItems {
			cloned := src
			cloned.Id = uuid.Nil // re-issue
			cloned.WeekId = target.Id

			if mode == "actual" {
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
