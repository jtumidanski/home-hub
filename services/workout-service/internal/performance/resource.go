package performance

import (
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

func writePerformance(w http.ResponseWriter, l logrus.FieldLogger, si jsonapi.ServerInformation, m Model, sets []SetModel) {
	server.MarshalResponse[RestModel](l)(w)(si)(map[string][]string{})(Transform(m, sets))
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
				Status:                input.Status,
				WeightUnit:            input.WeightUnit,
				ActualSets:            input.ActualSets,
				ActualReps:            input.ActualReps,
				ActualWeight:          input.ActualWeight,
				ActualDurationSeconds: input.ActualDurationSeconds,
				ActualDistance:        input.ActualDistance,
				ActualDistanceUnit:    input.ActualDistanceUnit,
				Notes:                 input.Notes,
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, sets, err := proc.Patch(t.Id(), t.UserId(), itemID, in)
			if err != nil {
				writePerformanceError(w, d.Logger(), "Failed to patch performance", err)
				return
			}
			writePerformance(w, d.Logger(), c.ServerInformation(), m, sets)
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
			writePerformance(w, d.Logger(), c.ServerInformation(), m, sets)
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
			writePerformance(w, d.Logger(), c.ServerInformation(), m, nil)
		}
	}
}
