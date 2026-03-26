package invitation

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)

		api.HandleFunc("/invitations", rh("ListInvitations", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/invitations/mine", rh("ListMyInvitations", listMineHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/invitations", rihCreate("CreateInvitation", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/invitations/{id}", rh("RevokeInvitation", revokeHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/invitations/{id}/accept", rh("AcceptInvitation", acceptHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/invitations/{id}/decline", rh("DeclineInvitation", declineHandler(db))).Methods(http.MethodPost)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			householdIDStr := r.URL.Query().Get("filter[householdId]")
			if householdIDStr == "" {
				server.WriteError(w, http.StatusBadRequest, "Missing Filter", "filter[householdId] is required")
				return
			}
			householdID, err := uuid.Parse(householdIDStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Filter", "filter[householdId] must be a valid UUID")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.ByHouseholdPendingProvider(householdID)()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list invitations")
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

// listMineHandler returns pending invitations for the current user's email.
// Bypasses tenant filtering since the user may not have a tenant yet.
func listMineHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, ok := sharedauth.ClaimsFromContext(r.Context())
			if !ok || claims.Email == "" {
				server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "email not available in token")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.ByEmailPendingProvider(claims.Email)()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list user invitations")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST models")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Include household resources so the join flow can display household names.
			// Bypass tenant filter — the user may not have a tenant yet.
			noTenantCtx := database.WithoutTenantFilter(r.Context())
			mineRest := make([]MineRestModel, len(rest))
			for i, rm := range rest {
				mineRest[i] = MineRestModel{RestModel: rm}
				hhProc := household.NewProcessor(d.Logger(), noTenantCtx, db)
				hh, err := hhProc.ByIDProvider(rm.HouseholdID)()
				if err == nil {
					hhRest, err := household.Transform(hh)
					if err == nil {
						mineRest[i].Household = &hhRest
					}
				}
			}

			server.MarshalSliceResponse[MineRestModel](d.Logger())(w)(c.ServerInformation())(mineRest)
		}
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), input.HouseholdID, input.Email, input.Role, t.UserId())
			if err != nil {
				if errors.Is(err, ErrNotAuthorized) {
					server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
					return
				}
				if errors.Is(err, ErrAlreadyInvited) {
					server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
					return
				}
				if errors.Is(err, ErrAlreadyMember) || errors.Is(err, ErrInvalidRole) {
					server.WriteError(w, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
					return
				}
				if errors.Is(err, ErrEmailRequired) || errors.Is(err, ErrHouseholdIDRequired) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create invitation")
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

func revokeHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				err := proc.Revoke(id, t.UserId())
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrNotPending) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					if errors.Is(err, ErrNotAuthorized) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to revoke invitation")
					server.WriteError(w, http.StatusInternalServerError, "Revoke Failed", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func acceptHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				claims, ok := sharedauth.ClaimsFromContext(r.Context())
				if !ok || claims.Email == "" {
					server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "email not available in token")
					return
				}

				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Accept(id, t.UserId(), claims.Email, t.Id())
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrNotPending) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					if errors.Is(err, ErrExpired) {
						server.WriteError(w, http.StatusGone, "Gone", "invitation has expired")
						return
					}
					if errors.Is(err, ErrEmailMismatch) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					if errors.Is(err, ErrCrossTenant) || errors.Is(err, ErrAlreadyHasMembership) {
						server.WriteError(w, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to accept invitation")
					server.WriteError(w, http.StatusInternalServerError, "Accept Failed", "")
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

func declineHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				claims, ok := sharedauth.ClaimsFromContext(r.Context())
				if !ok || claims.Email == "" {
					server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "email not available in token")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Decline(id, claims.Email)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrNotPending) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					if errors.Is(err, ErrEmailMismatch) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to decline invitation")
					server.WriteError(w, http.StatusInternalServerError, "Decline Failed", "")
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
