package category

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

		api.HandleFunc("/ingredient-categories", rh("ListCategories", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/ingredient-categories", rihCreate("CreateCategory", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/ingredient-categories/{id}", rihUpdate("UpdateCategory", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/ingredient-categories/{id}", rh("DeleteCategory", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)

			models, err := proc.List(t.Id())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list ingredient categories")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest := make([]*RestModel, len(models))
			for i, m := range models {
				r := Transform(m)
				rest[i] = &r
			}

			server.MarshalSliceResponse[*RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)

			m, err := proc.Create(t.Id(), input.Name)
			if err != nil {
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) {
					server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", err.Error())
					return
				}
				if errors.Is(err, ErrDuplicateName) {
					server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create ingredient category")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest := Transform(m)
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)

				var name *string
				if input.Name != "" {
					name = &input.Name
				}

				m, err := proc.Update(id, t.Id(), name, input.SortOrder)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Category not found")
						return
					}
					if errors.Is(err, ErrDuplicateName) {
						server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
						return
					}
					if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) || errors.Is(err, ErrInvalidSortOrder) {
						server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to update ingredient category")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest := Transform(m)
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)

				if err := proc.Delete(id, t.Id()); err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Category not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete ingredient category")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
