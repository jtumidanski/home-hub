package performance

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes mounts the performance write endpoints. PATCH and PUT use
// `RegisterInputHandler[T]` so api2go handles the JSON:API envelope; DELETE
// has no body so it stays on the GET handler signature.
func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihPatch := server.RegisterInputHandler[PatchPerformanceRequest](l)(si)
		rihPutSets := server.RegisterInputHandler[PutPerformanceSetsRequest](l)(si)

		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}/performance", rihPatch("PatchPerformance", patchHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}/performance/sets", rihPutSets("PutPerformanceSets", putSetsHandler(db))).Methods(http.MethodPut)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}/performance/sets", rh("DeletePerformanceSets", deleteSetsHandler(db))).Methods(http.MethodDelete)
	}
}

func parseItemID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(mux.Vars(r)["itemId"])
}

type performanceRest struct {
	Status     string       `json:"status"`
	Mode       string       `json:"mode"`
	WeightUnit *string      `json:"weightUnit,omitempty"`
	Actuals    *actualsRest `json:"actuals,omitempty"`
	Sets       []setRest    `json:"sets,omitempty"`
	Notes      *string      `json:"notes,omitempty"`
}

type actualsRest struct {
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	Weight          *float64 `json:"weight,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	Distance        *float64 `json:"distance,omitempty"`
	DistanceUnit    *string  `json:"distanceUnit,omitempty"`
}

type setRest struct {
	SetNumber int     `json:"setNumber"`
	Reps      int     `json:"reps"`
	Weight    float64 `json:"weight"`
}

func projectPerformance(m Model, sets []SetModel) performanceRest {
	out := performanceRest{
		Status:     m.Status(),
		Mode:       m.Mode(),
		WeightUnit: m.WeightUnit(),
		Notes:      m.Notes(),
	}
	if m.Mode() == ModePerSet {
		rows := make([]setRest, 0, len(sets))
		for _, s := range sets {
			rows = append(rows, setRest{SetNumber: s.SetNumber(), Reps: s.Reps(), Weight: s.Weight()})
		}
		out.Sets = rows
	} else {
		out.Actuals = &actualsRest{
			Sets:            m.ActualSets(),
			Reps:            m.ActualReps(),
			Weight:          m.ActualWeight(),
			DurationSeconds: m.ActualDurationSeconds(),
			Distance:        m.ActualDistance(),
			DistanceUnit:    m.ActualDistanceUnit(),
		}
	}
	return out
}

func writePerformance(w http.ResponseWriter, status int, m Model, sets []SetModel) {
	envelope := struct {
		Data struct {
			Type       string          `json:"type"`
			ID         string          `json:"id"`
			Attributes performanceRest `json:"attributes"`
		} `json:"data"`
	}{}
	envelope.Data.Type = "performances"
	envelope.Data.ID = m.PlannedItemID().String()
	envelope.Data.Attributes = projectPerformance(m, sets)
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope)
}

func writePerformanceError(w http.ResponseWriter, l logrus.FieldLogger, op string, err error) {
	switch {
	case errors.Is(err, ErrPlannedItemNotFound):
		server.WriteError(w, http.StatusNotFound, "Not Found", "Planned item not found")
	case errors.Is(err, ErrPerSetNotAllowed):
		server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", err.Error())
	case errors.Is(err, ErrSummaryWhilePerSet), errors.Is(err, ErrUnitChangeWithSets):
		server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
	case errors.Is(err, ErrInvalidStatus), errors.Is(err, ErrInvalidMode), errors.Is(err, ErrInvalidWeightUnit), errors.Is(err, ErrInvalidNumeric), errors.Is(err, ErrInvalidSetNumeric):
		server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
	default:
		l.WithError(err).Error(op)
		server.WriteError(w, http.StatusInternalServerError, "Error", "")
	}
}

func patchHandler(db *gorm.DB) server.InputHandler[PatchPerformanceRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input PatchPerformanceRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			itemID, err := parseItemID(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid itemId", "itemId must be a UUID")
				return
			}

			in := PatchInput{
				Status:     input.Status,
				WeightUnit: input.WeightUnit,
				Notes:      input.Notes,
			}
			if input.Actuals != nil {
				in.ActualSets = input.Actuals.Sets
				in.ActualReps = input.Actuals.Reps
				in.ActualWeight = input.Actuals.Weight
				in.ActualDurationSeconds = input.Actuals.DurationSeconds
				in.ActualDistance = input.Actuals.Distance
				in.ActualDistanceUnit = input.Actuals.DistanceUnit
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, sets, err := proc.Patch(t.Id(), t.UserId(), itemID, in)
			if err != nil {
				writePerformanceError(w, d.Logger(), "Failed to patch performance", err)
				return
			}
			writePerformance(w, http.StatusOK, m, sets)
		}
	}
}

func putSetsHandler(db *gorm.DB) server.InputHandler[PutPerformanceSetsRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input PutPerformanceSetsRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			itemID, err := parseItemID(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid itemId", "itemId must be a UUID")
				return
			}
			inputs := make([]SetInput, len(input.Sets))
			for i, s := range input.Sets {
				inputs[i] = SetInput{Reps: s.Reps, Weight: s.Weight}
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, sets, err := proc.ReplaceSets(t.Id(), t.UserId(), itemID, input.WeightUnit, inputs)
			if err != nil {
				writePerformanceError(w, d.Logger(), "Failed to replace performance sets", err)
				return
			}
			writePerformance(w, http.StatusOK, m, sets)
		}
	}
}

func deleteSetsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			itemID, err := parseItemID(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid itemId", "itemId must be a UUID")
				return
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.CollapseSets(t.UserId(), itemID)
			if err != nil {
				writePerformanceError(w, d.Logger(), "Failed to collapse performance sets", err)
				return
			}
			writePerformance(w, http.StatusOK, m, nil)
		}
	}
}
