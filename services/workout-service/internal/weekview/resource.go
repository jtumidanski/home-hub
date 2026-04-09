package weekview

import (
	"encoding/json"
	"errors"
	"io"
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

// InitializeRoutes mounts both the week endpoints and the planned-item
// endpoints under /workouts/weeks/{weekStart}/...
func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		// Order: more specific paths first so the weekStart-only routes don't
		// swallow the items/* and items/{itemId} variants.
		api.HandleFunc("/workouts/weeks/{weekStart}/copy", rh("CopyWeek", CopyHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/bulk", rh("BulkAddPlannedItems", bulkAddHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/reorder", rh("ReorderPlannedItems", reorderHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}", rh("UpdatePlannedItem", updateItemHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}", rh("DeletePlannedItem", deleteItemHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/workouts/weeks/{weekStart}/items", rh("AddPlannedItem", addItemHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/weeks/{weekStart}", rh("GetWeek", getWeekHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/workouts/weeks/{weekStart}", rh("PatchWeek", patchWeekHandler(db))).Methods(http.MethodPatch)
	}
}

func parseWeekStart(r *http.Request) (time.Time, error) {
	return time.ParseInLocation(weekDateLayout, mux.Vars(r)["weekStart"], time.UTC)
}

func parseItemID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(mux.Vars(r)["itemId"])
}

// renderWeek loads the week and its embedded items, marshals the JSON:API
// document, and writes the response. Used by every mutation endpoint so the
// client gets the post-update state in one round-trip.
func renderWeek(w http.ResponseWriter, l logrus.FieldLogger, weekProc *week.Processor, weekModel week.Model, status int) {
	items, err := AssembleItems(weekProc.DB(), weekModel.Id())
	if err != nil {
		l.WithError(err).Error("Failed to assemble week items")
		server.WriteError(w, http.StatusInternalServerError, "Error", "")
		return
	}
	body, err := MarshalDocument(BuildDocument(weekModel, items))
	if err != nil {
		l.WithError(err).Error("Failed to marshal week document")
		server.WriteError(w, http.StatusInternalServerError, "Error", "")
		return
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func ensureAndRender(w http.ResponseWriter, l logrus.FieldLogger, weekProc *week.Processor, tenantID, userID uuid.UUID, weekStart time.Time, status int) bool {
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
	renderWeek(w, l, weekProc, m, status)
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
			renderWeek(w, d.Logger(), weekProc, m, http.StatusOK)
		}
	}
}

func patchWeekHandler(db *gorm.DB) server.GetHandler {
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
						RestDayFlags *[]int `json:"restDayFlags"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}
			if env.Data.Attributes.RestDayFlags == nil {
				server.WriteError(w, http.StatusBadRequest, "Validation Failed", "restDayFlags is required")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			m, err := weekProc.PatchRestDayFlags(t.Id(), t.UserId(), weekStart, *env.Data.Attributes.RestDayFlags)
			if err != nil {
				if errors.Is(err, week.ErrInvalidRestDay) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to patch week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			renderWeek(w, d.Logger(), weekProc, m, http.StatusOK)
		}
	}
}

// itemAttrs is the per-item create payload shared by single and bulk add.
type itemAttrs struct {
	ExerciseID uuid.UUID `json:"exerciseId"`
	DayOfWeek  int       `json:"dayOfWeek"`
	Position   *int      `json:"position,omitempty"`
	Planned    *struct {
		Sets            *int     `json:"sets"`
		Reps            *int     `json:"reps"`
		Weight          *float64 `json:"weight"`
		WeightUnit      *string  `json:"weightUnit"`
		DurationSeconds *int     `json:"durationSeconds"`
		Distance        *float64 `json:"distance"`
		DistanceUnit    *string  `json:"distanceUnit"`
	} `json:"planned,omitempty"`
	Notes *string `json:"notes,omitempty"`
}

func (a itemAttrs) toAddInput() planneditem.AddInput {
	in := planneditem.AddInput{
		ExerciseID: a.ExerciseID,
		DayOfWeek:  a.DayOfWeek,
		Position:   a.Position,
		Notes:      a.Notes,
	}
	if a.Planned != nil {
		in.PlannedSets = a.Planned.Sets
		in.PlannedReps = a.Planned.Reps
		in.PlannedWeight = a.Planned.Weight
		in.PlannedWeightUnit = a.Planned.WeightUnit
		in.PlannedDurationSeconds = a.Planned.DurationSeconds
		in.PlannedDistance = a.Planned.Distance
		in.PlannedDistanceUnit = a.Planned.DistanceUnit
	}
	return in
}

func addItemHandler(db *gorm.DB) server.GetHandler {
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
					Attributes itemAttrs `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			weekEntity, err := weekProc.EnsureExists(t.Id(), t.UserId(), weekStart)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to ensure week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			proc := planneditem.NewProcessor(d.Logger(), r.Context(), db)
			if _, err := proc.Add(t.Id(), t.UserId(), weekEntity.Id, env.Data.Attributes.toAddInput()); err != nil {
				writePlannedItemError(w, d.Logger(), "Failed to add planned item", err)
				return
			}
			ensureAndRender(w, d.Logger(), weekProc, t.Id(), t.UserId(), weekStart, http.StatusCreated)
		}
	}
}

func bulkAddHandler(db *gorm.DB) server.GetHandler {
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
						Items []itemAttrs `json:"items"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			weekEntity, err := weekProc.EnsureExists(t.Id(), t.UserId(), weekStart)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to ensure week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			inputs := make([]planneditem.AddInput, 0, len(env.Data.Attributes.Items))
			for _, a := range env.Data.Attributes.Items {
				inputs = append(inputs, a.toAddInput())
			}
			proc := planneditem.NewProcessor(d.Logger(), r.Context(), db)
			if _, err := proc.BulkAdd(t.Id(), t.UserId(), weekEntity.Id, inputs); err != nil {
				writePlannedItemError(w, d.Logger(), "Failed to bulk add planned items", err)
				return
			}
			ensureAndRender(w, d.Logger(), weekProc, t.Id(), t.UserId(), weekStart, http.StatusCreated)
		}
	}
}

func updateItemHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
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
			body, _ := io.ReadAll(r.Body)
			var env struct {
				Data struct {
					Attributes struct {
						DayOfWeek *int `json:"dayOfWeek,omitempty"`
						Position  *int `json:"position,omitempty"`
						Planned   *struct {
							Sets            *int     `json:"sets"`
							Reps            *int     `json:"reps"`
							Weight          *float64 `json:"weight"`
							WeightUnit      *string  `json:"weightUnit"`
							DurationSeconds *int     `json:"durationSeconds"`
							Distance        *float64 `json:"distance"`
							DistanceUnit    *string  `json:"distanceUnit"`
						} `json:"planned,omitempty"`
						Notes *string `json:"notes,omitempty"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}
			ui := planneditem.UpdateInput{
				DayOfWeek: env.Data.Attributes.DayOfWeek,
				Position:  env.Data.Attributes.Position,
				Notes:     env.Data.Attributes.Notes,
			}
			if env.Data.Attributes.Planned != nil {
				ui.PlannedSets = env.Data.Attributes.Planned.Sets
				ui.PlannedReps = env.Data.Attributes.Planned.Reps
				ui.PlannedWeight = env.Data.Attributes.Planned.Weight
				ui.PlannedWeightUnit = env.Data.Attributes.Planned.WeightUnit
				ui.PlannedDurationSeconds = env.Data.Attributes.Planned.DurationSeconds
				ui.PlannedDistance = env.Data.Attributes.Planned.Distance
				ui.PlannedDistanceUnit = env.Data.Attributes.Planned.DistanceUnit
			}
			proc := planneditem.NewProcessor(d.Logger(), r.Context(), db)
			if _, err := proc.Update(itemID, ui); err != nil {
				writePlannedItemError(w, d.Logger(), "Failed to update planned item", err)
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			ensureAndRender(w, d.Logger(), weekProc, t.Id(), t.UserId(), weekStart, http.StatusOK)
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

func reorderHandler(db *gorm.DB) server.GetHandler {
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
						Items []struct {
							ItemID    uuid.UUID `json:"itemId"`
							DayOfWeek int       `json:"dayOfWeek"`
							Position  int       `json:"position"`
						} `json:"items"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}
			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			weekEntity, err := weekProc.EnsureExists(t.Id(), t.UserId(), weekStart)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to ensure week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			entries := make([]planneditem.ReorderEntry, 0, len(env.Data.Attributes.Items))
			for _, it := range env.Data.Attributes.Items {
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
			ensureAndRender(w, d.Logger(), weekProc, t.Id(), t.UserId(), weekStart, http.StatusOK)
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
