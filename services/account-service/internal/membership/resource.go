package membership

import (
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

		api.HandleFunc("/memberships", rh("ListMemberships", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/memberships", rihCreate("CreateMembership", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/memberships/{id}", rihUpdate("UpdateMembership", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/memberships/{id}", rh("DeleteMembership", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.ByUserProvider(t.UserId())()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list memberships")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST models")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), input.HouseholdId, input.UserId, input.Role)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to create membership")
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}
			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			server.ParseID(r, w, "id", func(id uuid.UUID) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					proc := NewProcessor(d.Logger(), r.Context(), db)
					m, err := proc.UpdateRole(id, input.Role)
					if err != nil {
						d.Logger().WithError(err).Error("Failed to update membership")
						server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
						return
					}
					rest, err := Transform(m)
					if err != nil {
						d.Logger().WithError(err).Error("Creating REST model")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
				}
			})(w, r)
		}
	}
}

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			server.ParseID(r, w, "id", func(id uuid.UUID) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					proc := NewProcessor(d.Logger(), r.Context(), db)
					if err := proc.Delete(id); err != nil {
						d.Logger().WithError(err).Error("Failed to delete membership")
						server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
						return
					}
					w.WriteHeader(http.StatusNoContent)
				}
			})(w, r)
		}
	}
}
