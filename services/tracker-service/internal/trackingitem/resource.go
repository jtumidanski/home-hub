package trackingitem

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

		api.HandleFunc("/trackers", rh("ListTrackers", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/trackers", rihCreate("CreateTracker", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/trackers/{id}", rh("GetTracker", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/trackers/{id}", rihUpdate("UpdateTracker", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/trackers/{id}", rh("DeleteTracker", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)

			rows, err := proc.ListWithSchedules(t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list trackers")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest := TransformSlice(rows)
			server.MarshalSliceResponse[*RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)

				m, err := proc.Get(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Tracking item not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to get tracker")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				sched, _ := proc.GetCurrentSchedule(m.Id())
				history, _ := proc.GetScheduleHistory(m.Id())
				rest := Transform(m, sched, history)
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)

			m, err := proc.Create(t.Id(), t.UserId(), input.Name, input.ScaleType, input.Color, input.ScaleConfig, input.Schedule, input.SortOrder)
			if err != nil {
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) ||
					errors.Is(err, ErrInvalidScaleType) || errors.Is(err, ErrInvalidColor) ||
					errors.Is(err, ErrInvalidSortOrder) || errors.Is(err, ErrRangeConfigRequired) ||
					errors.Is(err, ErrInvalidRangeConfig) || errors.Is(err, ErrInvalidScheduleDay) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				if errors.Is(err, ErrDuplicateName) {
					server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create tracker")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			sched, _ := proc.GetCurrentSchedule(m.Id())
			history, _ := proc.GetScheduleHistory(m.Id())
			rest := Transform(m, sched, history)
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if input.ScaleType != "" {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", ErrScaleTypeImmutable.Error())
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)

				var name *string
				if input.Name != "" {
					name = &input.Name
				}
				var color *string
				if input.Color != "" {
					color = &input.Color
				}

				m, err := proc.Update(id, name, color, input.Schedule, input.SortOrder, input.ScaleConfig)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Tracking item not found")
						return
					}
					if errors.Is(err, ErrDuplicateName) {
						server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
						return
					}
					if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) ||
						errors.Is(err, ErrInvalidColor) || errors.Is(err, ErrInvalidSortOrder) ||
						errors.Is(err, ErrRangeConfigRequired) || errors.Is(err, ErrInvalidRangeConfig) ||
						errors.Is(err, ErrInvalidScheduleDay) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to update tracker")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				sched, _ := proc.GetCurrentSchedule(m.Id())
				history, _ := proc.GetScheduleHistory(m.Id())
				rest := Transform(m, sched, history)
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
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
						server.WriteError(w, http.StatusNotFound, "Not Found", "Tracking item not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete tracker")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
