package weekview

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const weekDateLayout = "2006-01-02"

// InitializeRoutes mounts the week endpoints and the planned-item endpoints
// under `/workouts/weeks/{weekStart}/...`. Write endpoints use
// `RegisterInputHandler[T]` so api2go handles the JSON:API envelope; only
// pure GETs (and the `DELETE .../items/{itemId}`) use `RegisterHandler`.
func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihPatchWeek := server.RegisterInputHandler[PatchWeekRequest](l)(si)
		rihCopy := server.RegisterInputHandler[CopyWeekRequest](l)(si)
		rihAdd := server.RegisterInputHandler[AddPlannedItemRequest](l)(si)
		rihBulk := server.RegisterInputHandler[BulkAddPlannedItemsRequest](l)(si)
		rihReorder := server.RegisterInputHandler[ReorderPlannedItemsRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdatePlannedItemRequest](l)(si)

		// Order: more specific paths first so the weekStart-only routes don't
		// swallow the items/* and items/{itemId} variants. The `/nearest`
		// literal also needs to register ahead of `{weekStart}` so the router
		// matches the helper endpoint rather than trying to parse "nearest"
		// as a date.
		api.HandleFunc("/workouts/weeks/nearest", rh("GetNearestPopulatedWeek", nearestHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/workouts/weeks/{weekStart}/copy", rihCopy("CopyWeek", copyHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/bulk", rihBulk("BulkAddPlannedItems", bulkAddHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/reorder", rihReorder("ReorderPlannedItems", reorderHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}", rihUpdate("UpdatePlannedItem", updateItemHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}", rh("DeletePlannedItem", deleteItemHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/workouts/weeks/{weekStart}/items", rihAdd("AddPlannedItem", addItemHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/weeks/{weekStart}", rh("GetWeek", getWeekHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/workouts/weeks/{weekStart}", rihPatchWeek("PatchWeek", patchWeekHandler(db))).Methods(http.MethodPatch)
	}
}

func parseWeekStart(r *http.Request) (time.Time, error) {
	return time.ParseInLocation(weekDateLayout, mux.Vars(r)["weekStart"], time.UTC)
}

func parseItemID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(mux.Vars(r)["itemId"])
}

// renderWeek loads a week's embedded items via the weekview processor and
// writes the JSON:API envelope. Used by every mutation endpoint so the client
// gets the post-update state in one round-trip.
func renderWeek(w http.ResponseWriter, l logrus.FieldLogger, si jsonapi.ServerInformation, viewProc *Processor, weekModel week.Model, status int) {
	rm, err := viewProc.LoadWeekDocument(weekModel)
	if err != nil {
		l.WithError(err).Error("Failed to load week document")
		server.WriteError(w, http.StatusInternalServerError, "Error", "")
		return
	}
	if status == http.StatusCreated {
		server.MarshalCreatedResponse[RestModel](l)(w)(si)(rm)
		return
	}
	server.MarshalResponse[RestModel](l)(w)(si)(map[string][]string{})(rm)
}

// ensureAndRender lazily creates the week if missing, then renders the
// post-update state. Mutation handlers call this after the underlying domain
// processor has applied its change.
func ensureAndRender(w http.ResponseWriter, l logrus.FieldLogger, si jsonapi.ServerInformation, db *gorm.DB, ctx context.Context, tenantID, userID uuid.UUID, weekStart time.Time, status int) bool {
	weekProc := week.NewProcessor(l, ctx, db)
	e, err := weekProc.EnsureExists(tenantID, userID, weekStart)
	if err != nil {
		l.WithError(err).Error("Failed to ensure week")
		server.WriteError(w, http.StatusInternalServerError, "Error", "")
		return false
	}
	m, err := week.Make(e)
	if err != nil {
		l.WithError(err).Error("Failed to materialize week")
		server.WriteError(w, http.StatusInternalServerError, "Error", "")
		return false
	}
	viewProc := NewProcessor(l, ctx, db)
	renderWeek(w, l, si, viewProc, m, status)
	return true
}

func getWeekHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			weekStart, err := parseWeekStart(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			m, err := weekProc.Get(t.UserId(), weekStart)
			if err != nil {
				if errors.Is(err, week.ErrNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Week not found")
					return
				}
				d.Logger().WithError(err).Error("Failed to load week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			viewProc := NewProcessor(d.Logger(), r.Context(), db)
			renderWeek(w, d.Logger(), c.ServerInformation(), viewProc, m, http.StatusOK)
		}
	}
}

func patchWeekHandler(db *gorm.DB) server.InputHandler[PatchWeekRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input PatchWeekRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			weekStart, err := parseWeekStart(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}
			if input.RestDayFlags == nil {
				server.WriteError(w, http.StatusBadRequest, "Validation Failed", "restDayFlags is required")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			m, err := weekProc.PatchRestDayFlags(t.Id(), t.UserId(), weekStart, *input.RestDayFlags)
			if err != nil {
				if errors.Is(err, week.ErrInvalidRestDay) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to patch week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			viewProc := NewProcessor(d.Logger(), r.Context(), db)
			renderWeek(w, d.Logger(), c.ServerInformation(), viewProc, m, http.StatusOK)
		}
	}
}

// plannedFields groups the flat planned-values fields shared by add and bulk-add
// request structs. Extracted so toAddInput can accept either caller shape.
type plannedFields struct {
	PlannedSets            *int
	PlannedReps            *int
	PlannedWeight          *float64
	PlannedWeightUnit      *string
	PlannedDurationSeconds *int
	PlannedDistance        *float64
	PlannedDistanceUnit    *string
}

// toAddInput projects flat planned-item attrs onto the processor input.
// Shared by single-add and bulk-add handlers.
func toAddInput(exerciseID uuid.UUID, dayOfWeek int, position *int, pf plannedFields, notes *string) planneditem.AddInput {
	return planneditem.AddInput{
		ExerciseID:             exerciseID,
		DayOfWeek:              dayOfWeek,
		Position:               position,
		PlannedSets:            pf.PlannedSets,
		PlannedReps:            pf.PlannedReps,
		PlannedWeight:          pf.PlannedWeight,
		PlannedWeightUnit:      pf.PlannedWeightUnit,
		PlannedDurationSeconds: pf.PlannedDurationSeconds,
		PlannedDistance:        pf.PlannedDistance,
		PlannedDistanceUnit:    pf.PlannedDistanceUnit,
		Notes:                  notes,
	}
}

func addItemHandler(db *gorm.DB) server.InputHandler[AddPlannedItemRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input AddPlannedItemRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			weekStart, err := parseWeekStart(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			weekEntity, err := weekProc.EnsureExists(t.Id(), t.UserId(), weekStart)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to ensure week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			in := toAddInput(input.ExerciseID, input.DayOfWeek, input.Position, plannedFields{
				PlannedSets:            input.PlannedSets,
				PlannedReps:            input.PlannedReps,
				PlannedWeight:          input.PlannedWeight,
				PlannedWeightUnit:      input.PlannedWeightUnit,
				PlannedDurationSeconds: input.PlannedDurationSeconds,
				PlannedDistance:        input.PlannedDistance,
				PlannedDistanceUnit:    input.PlannedDistanceUnit,
			}, input.Notes)
			proc := planneditem.NewProcessor(d.Logger(), r.Context(), db)
			if _, err := proc.Add(t.Id(), t.UserId(), weekEntity.Id, in); err != nil {
				writePlannedItemError(w, d.Logger(), "Failed to add planned item", err)
				return
			}
			ensureAndRender(w, d.Logger(), c.ServerInformation(), db, r.Context(), t.Id(), t.UserId(), weekStart, http.StatusCreated)
		}
	}
}

func bulkAddHandler(db *gorm.DB) server.InputHandler[BulkAddPlannedItemsRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input BulkAddPlannedItemsRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			weekStart, err := parseWeekStart(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			weekEntity, err := weekProc.EnsureExists(t.Id(), t.UserId(), weekStart)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to ensure week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			inputs := make([]planneditem.AddInput, 0, len(input.Items))
			for _, a := range input.Items {
				inputs = append(inputs, toAddInput(a.ExerciseID, a.DayOfWeek, a.Position, plannedFields{
					PlannedSets:            a.PlannedSets,
					PlannedReps:            a.PlannedReps,
					PlannedWeight:          a.PlannedWeight,
					PlannedWeightUnit:      a.PlannedWeightUnit,
					PlannedDurationSeconds: a.PlannedDurationSeconds,
					PlannedDistance:        a.PlannedDistance,
					PlannedDistanceUnit:    a.PlannedDistanceUnit,
				}, a.Notes))
			}
			proc := planneditem.NewProcessor(d.Logger(), r.Context(), db)
			if _, err := proc.BulkAdd(t.Id(), t.UserId(), weekEntity.Id, inputs); err != nil {
				writePlannedItemError(w, d.Logger(), "Failed to bulk add planned items", err)
				return
			}
			ensureAndRender(w, d.Logger(), c.ServerInformation(), db, r.Context(), t.Id(), t.UserId(), weekStart, http.StatusCreated)
		}
	}
}

func updateItemHandler(db *gorm.DB) server.InputHandler[UpdatePlannedItemRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdatePlannedItemRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			weekStart, err := parseWeekStart(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}
			itemID, err := parseItemID(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid itemId", "itemId must be a UUID")
				return
			}
			ui := planneditem.UpdateInput{
				DayOfWeek:              input.DayOfWeek,
				Position:               input.Position,
				PlannedSets:            input.PlannedSets,
				PlannedReps:            input.PlannedReps,
				PlannedWeight:          input.PlannedWeight,
				PlannedWeightUnit:      input.PlannedWeightUnit,
				PlannedDurationSeconds: input.PlannedDurationSeconds,
				PlannedDistance:        input.PlannedDistance,
				PlannedDistanceUnit:    input.PlannedDistanceUnit,
				Notes:                  input.Notes,
			}
			proc := planneditem.NewProcessor(d.Logger(), r.Context(), db)
			if _, err := proc.Update(itemID, ui); err != nil {
				writePlannedItemError(w, d.Logger(), "Failed to update planned item", err)
				return
			}
			ensureAndRender(w, d.Logger(), c.ServerInformation(), db, r.Context(), t.Id(), t.UserId(), weekStart, http.StatusOK)
		}
	}
}

func deleteItemHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			itemID, err := parseItemID(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid itemId", "itemId must be a UUID")
				return
			}
			proc := planneditem.NewProcessor(d.Logger(), r.Context(), db)
			if err := proc.Delete(itemID); err != nil {
				writePlannedItemError(w, d.Logger(), "Failed to delete planned item", err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func reorderHandler(db *gorm.DB) server.InputHandler[ReorderPlannedItemsRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input ReorderPlannedItemsRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			weekStart, err := parseWeekStart(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			weekEntity, err := weekProc.EnsureExists(t.Id(), t.UserId(), weekStart)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to ensure week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			entries := make([]planneditem.ReorderEntry, 0, len(input.Items))
			for _, it := range input.Items {
				entries = append(entries, planneditem.ReorderEntry{
					ItemID:    it.ItemID,
					DayOfWeek: it.DayOfWeek,
					Position:  it.Position,
				})
			}
			proc := planneditem.NewProcessor(d.Logger(), r.Context(), db)
			if err := proc.Reorder(weekEntity.Id, entries); err != nil {
				writePlannedItemError(w, d.Logger(), "Failed to reorder planned items", err)
				return
			}
			ensureAndRender(w, d.Logger(), c.ServerInformation(), db, r.Context(), t.Id(), t.UserId(), weekStart, http.StatusOK)
		}
	}
}

func copyHandler(db *gorm.DB) server.InputHandler[CopyWeekRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CopyWeekRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			weekStart, err := parseWeekStart(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}
			viewProc := NewProcessor(d.Logger(), r.Context(), db)
			created, err := viewProc.Copy(t.Id(), t.UserId(), weekStart, input.Mode)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidCopyMode):
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
				case errors.Is(err, ErrCopyTargetNotEmpty):
					server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
				case errors.Is(err, ErrCopyNoSource):
					server.WriteError(w, http.StatusNotFound, "Not Found", err.Error())
				default:
					d.Logger().WithError(err).Error("Failed to copy week")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
				}
				return
			}
			renderWeek(w, d.Logger(), c.ServerInformation(), viewProc, created, http.StatusCreated)
		}
	}
}

func writePlannedItemError(w http.ResponseWriter, l logrus.FieldLogger, op string, err error) {
	switch {
	case errors.Is(err, planneditem.ErrNotFound):
		server.WriteError(w, http.StatusNotFound, "Not Found", "Planned item not found")
	case errors.Is(err, planneditem.ErrExerciseNotFound), errors.Is(err, planneditem.ErrExerciseMismatch):
		server.WriteError(w, http.StatusNotFound, "Not Found", err.Error())
	case errors.Is(err, planneditem.ErrExerciseDeleted):
		server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", err.Error())
	case errors.Is(err, planneditem.ErrInvalidDayOfWeek), errors.Is(err, planneditem.ErrInvalidPosition), errors.Is(err, planneditem.ErrInvalidNumeric):
		server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
	default:
		l.WithError(err).Error(op)
		server.WriteError(w, http.StatusInternalServerError, "Error", "")
	}
}
