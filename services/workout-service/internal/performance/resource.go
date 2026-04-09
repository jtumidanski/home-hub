package performance

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}/performance", rh("PatchPerformance", patchHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}/performance/sets", rh("PutPerformanceSets", putSetsHandler(db))).Methods(http.MethodPut)
		api.HandleFunc("/workouts/weeks/{weekStart}/items/{itemId}/performance/sets", rh("DeletePerformanceSets", deleteSetsHandler(db))).Methods(http.MethodDelete)
	}
}

func parseItemID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(mux.Vars(r)["itemId"])
}

type performanceRest struct {
	Status     string                `json:"status"`
	Mode       string                `json:"mode"`
	WeightUnit *string               `json:"weightUnit,omitempty"`
	Actuals    *actualsRest          `json:"actuals,omitempty"`
	Sets       []setRest             `json:"sets,omitempty"`
	Notes      *string               `json:"notes,omitempty"`
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

func patchHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			itemID, err := parseItemID(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid itemId", "itemId must be a UUID")
				return
			}
			body, _ := io.ReadAll(r.Body)
			var env struct {
				Data struct {
					Attributes struct {
						Status     *string `json:"status,omitempty"`
						WeightUnit *string `json:"weightUnit,omitempty"`
						Actuals    *struct {
							Sets            *int     `json:"sets,omitempty"`
							Reps            *int     `json:"reps,omitempty"`
							Weight          *float64 `json:"weight,omitempty"`
							DurationSeconds *int     `json:"durationSeconds,omitempty"`
							Distance        *float64 `json:"distance,omitempty"`
							DistanceUnit    *string  `json:"distanceUnit,omitempty"`
						} `json:"actuals,omitempty"`
						Notes *string `json:"notes,omitempty"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}

			in := PatchInput{
				Status:     env.Data.Attributes.Status,
				WeightUnit: env.Data.Attributes.WeightUnit,
				Notes:      env.Data.Attributes.Notes,
			}
			if env.Data.Attributes.Actuals != nil {
				in.ActualSets = env.Data.Attributes.Actuals.Sets
				in.ActualReps = env.Data.Attributes.Actuals.Reps
				in.ActualWeight = env.Data.Attributes.Actuals.Weight
				in.ActualDurationSeconds = env.Data.Attributes.Actuals.DurationSeconds
				in.ActualDistance = env.Data.Attributes.Actuals.Distance
				in.ActualDistanceUnit = env.Data.Attributes.Actuals.DistanceUnit
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

func putSetsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			itemID, err := parseItemID(r)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid itemId", "itemId must be a UUID")
				return
			}
			body, _ := io.ReadAll(r.Body)
			var env struct {
				Data struct {
					Attributes struct {
						WeightUnit string `json:"weightUnit"`
						Sets       []struct {
							Reps   int     `json:"reps"`
							Weight float64 `json:"weight"`
						} `json:"sets"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &env); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}
			inputs := make([]SetInput, len(env.Data.Attributes.Sets))
			for i, s := range env.Data.Attributes.Sets {
				inputs[i] = SetInput{Reps: s.Reps, Weight: s.Weight}
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, sets, err := proc.ReplaceSets(t.Id(), t.UserId(), itemID, env.Data.Attributes.WeightUnit, inputs)
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
