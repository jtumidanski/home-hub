package exercise

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

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)

		api.HandleFunc("/workouts/exercises", rh("ListExercises", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/workouts/exercises", rihCreate("CreateExercise", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/exercises/{id}", rihUpdate("UpdateExercise", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/workouts/exercises/{id}", rh("DeleteExercise", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)

			var themeID, regionID *uuid.UUID
			if v := r.URL.Query().Get("themeId"); v != "" {
				if id, err := uuid.Parse(v); err == nil {
					themeID = &id
				} else {
					server.WriteError(w, http.StatusBadRequest, "Invalid themeId", "themeId must be a UUID")
					return
				}
			}
			if v := r.URL.Query().Get("regionId"); v != "" {
				if id, err := uuid.Parse(v); err == nil {
					regionID = &id
				} else {
					server.WriteError(w, http.StatusBadRequest, "Invalid regionId", "regionId must be a UUID")
					return
				}
			}

			models, err := proc.List(t.UserId(), themeID, regionID)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list exercises")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalSliceResponse[*RestModel](d.Logger())(w)(c.ServerInformation())(TransformSlice(models))
		}
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), t.UserId(), CreateInput{
				Name:               input.Name,
				Kind:               input.Kind,
				WeightType:         input.WeightType,
				ThemeID:            input.ThemeID,
				RegionID:           input.RegionID,
				SecondaryRegionIDs: input.SecondaryRegionIDs,
				Defaults: Defaults{
					Sets:            input.DefaultSets,
					Reps:            input.DefaultReps,
					Weight:          input.DefaultWeight,
					WeightUnit:      input.DefaultWeightUnit,
					DurationSeconds: input.DefaultDurationSeconds,
					Distance:        input.DefaultDistance,
					DistanceUnit:    input.DefaultDistanceUnit,
				},
				Notes: input.Notes,
			})
			if err != nil {
				writeExerciseError(w, d.Logger(), "Failed to create exercise", err)
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(Transform(m))
		}
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if input.Kind != "" {
					server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", ErrKindImmutable.Error())
					return
				}
				if input.WeightType != "" {
					server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", ErrWeightTypeImmutable.Error())
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				ui := UpdateInput{
					Name:               input.Name,
					ThemeID:            input.ThemeID,
					RegionID:           input.RegionID,
					SecondaryRegionIDs: input.SecondaryRegionIDs,
					Notes:              input.Notes,
				}
				if input.DefaultSets != nil || input.DefaultReps != nil || input.DefaultWeight != nil ||
					input.DefaultWeightUnit != nil || input.DefaultDurationSeconds != nil ||
					input.DefaultDistance != nil || input.DefaultDistanceUnit != nil {
					def := Defaults{
						Sets:            input.DefaultSets,
						Reps:            input.DefaultReps,
						Weight:          input.DefaultWeight,
						WeightUnit:      input.DefaultWeightUnit,
						DurationSeconds: input.DefaultDurationSeconds,
						Distance:        input.DefaultDistance,
						DistanceUnit:    input.DefaultDistanceUnit,
					}
					ui.Defaults = &def
				}
				m, err := proc.Update(id, ui)
				if err != nil {
					writeExerciseError(w, d.Logger(), "Failed to update exercise", err)
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(Transform(m))
			}
		})
	}
}

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				if err := proc.Delete(id); err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Exercise not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete exercise")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

// writeExerciseError centralizes the error → status mapping for the create and
// update handlers so the §3.7 contract (400/404/409/422) stays consistent.
func writeExerciseError(w http.ResponseWriter, l logrus.FieldLogger, op string, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		server.WriteError(w, http.StatusNotFound, "Not Found", "Exercise not found")
	case errors.Is(err, ErrThemeNotFound), errors.Is(err, ErrRegionNotFound), errors.Is(err, ErrSecondaryNotFound):
		server.WriteError(w, http.StatusNotFound, "Not Found", err.Error())
	case errors.Is(err, ErrDuplicateName):
		server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
	case errors.Is(err, ErrKindImmutable), errors.Is(err, ErrWeightTypeImmutable), errors.Is(err, ErrInvalidDefaultsShape):
		server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", err.Error())
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrNameTooLong),
		errors.Is(err, ErrInvalidKind), errors.Is(err, ErrInvalidWeightType),
		errors.Is(err, ErrInvalidWeightUnit), errors.Is(err, ErrInvalidDistanceUnit),
		errors.Is(err, ErrInvalidNumeric), errors.Is(err, ErrPrimaryInSecondary),
		errors.Is(err, ErrThemeRequired), errors.Is(err, ErrRegionRequired):
		server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
	default:
		l.WithError(err).Error(op)
		server.WriteError(w, http.StatusInternalServerError, "Error", "")
	}
}
