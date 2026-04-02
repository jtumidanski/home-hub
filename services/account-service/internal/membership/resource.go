package membership

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

			// Support filter[householdId] for listing household members
			householdIDStr := r.URL.Query().Get("filter[householdId]")
			if householdIDStr != "" {
				householdID, err := uuid.Parse(householdIDStr)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Filter", "filter[householdId] must be a valid UUID")
					return
				}
				models, err := proc.ByHouseholdProvider(householdID)()
				if err != nil {
					d.Logger().WithError(err).Error("Failed to list household memberships")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				// Compute isLastOwner
				ownerCount, err := proc.CountOwnersByHousehold(householdID)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to count owners")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest, err := TransformSlice(models)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST models")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				for i := range rest {
					if rest[i].Role == "owner" && ownerCount == 1 {
						rest[i].IsLastOwner = true
					}
				}
				server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
				return
			}

			// Default: list user's memberships
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
				if errors.Is(err, ErrHouseholdIDRequired) || errors.Is(err, ErrUserIDRequired) || errors.Is(err, ErrRoleRequired) {
					d.Logger().WithError(err).Error("Validation failed")
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
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
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.UpdateRoleAuthorized(id, input.Role, t.UserId())
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					if errors.Is(err, ErrRoleRequired) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					if errors.Is(err, ErrNotAuthorized) || errors.Is(err, ErrCannotModifySelf) || errors.Is(err, ErrCannotModifyOwner) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
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
		})
	}
}

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)

				if err := proc.DeleteAuthorizedWithCleanup(id, t.UserId()); err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					if errors.Is(err, ErrNotAuthorized) || errors.Is(err, ErrCannotRemoveOwner) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					if errors.Is(err, ErrLastOwner) {
						server.WriteError(w, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to delete membership")
					server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
