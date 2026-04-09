package region

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

		api.HandleFunc("/workouts/regions", rh("ListRegions", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/workouts/regions", rihCreate("CreateRegion", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/workouts/regions/{id}", rihUpdate("UpdateRegion", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/workouts/regions/{id}", rh("DeleteRegion", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.List(t.Id(), t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list regions")
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
			m, err := proc.Create(t.Id(), t.UserId(), input.Name, input.SortOrder)
			if err != nil {
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) || errors.Is(err, ErrInvalidSortOrder) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				if errors.Is(err, ErrDuplicateName) {
					server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create region")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
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
				proc := NewProcessor(d.Logger(), r.Context(), db)
				var name *string
				if input.Name != "" {
					name = &input.Name
				}
				m, err := proc.Update(id, name, input.SortOrder)
				if err != nil {
					switch {
					case errors.Is(err, ErrNotFound):
						server.WriteError(w, http.StatusNotFound, "Not Found", "Region not found")
					case errors.Is(err, ErrDuplicateName):
						server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
					case errors.Is(err, ErrNameRequired), errors.Is(err, ErrNameTooLong), errors.Is(err, ErrInvalidSortOrder):
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					default:
						d.Logger().WithError(err).Error("Failed to update region")
						server.WriteError(w, http.StatusInternalServerError, "Error", "")
					}
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
						server.WriteError(w, http.StatusNotFound, "Not Found", "Region not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete region")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
