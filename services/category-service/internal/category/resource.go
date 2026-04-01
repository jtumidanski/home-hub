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

		api.HandleFunc("/categories", rh("ListCategories", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/categories", rihCreate("CreateCategory", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/categories/{id}", rihUpdate("UpdateCategory", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/categories/{id}", rh("DeleteCategory", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)

			models, err := proc.List(t.Id())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list categories")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			transformed, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform categories")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest := make([]*RestModel, len(transformed))
			for i := range transformed {
				rest[i] = &transformed[i]
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
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				if errors.Is(err, ErrDuplicateName) {
					server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create category")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform category")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
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
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Category not found")
						return
					}
					if errors.Is(err, ErrDuplicateName) {
						server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
						return
					}
					if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) || errors.Is(err, ErrInvalidSortOrder) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to update category")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to transform category")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
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
						server.WriteError(w, http.StatusNotFound, "Not Found", "Category not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete category")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
