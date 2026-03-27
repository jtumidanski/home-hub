package connection

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/config"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/crypto"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/oauthstate"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SyncTrigger func(conn Model)
type CascadeDelete func(ctx context.Context, connectionID uuid.UUID)

func InitializeRoutes(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, cfg config.Config, syncTrigger SyncTrigger, cascadeDelete CascadeDelete) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rih := server.RegisterInputHandler[AuthorizeRequest](l)(si)

		api.HandleFunc("/calendar/connections", rh("ListConnections", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/calendar/connections/google/authorize", rih("AuthorizeGoogle", authorizeHandler(db, gcClient, cfg))).Methods(http.MethodPost)
		api.HandleFunc("/calendar/connections/{id}", rh("DeleteConnection", deleteHandler(db, gcClient, enc, cascadeDelete))).Methods(http.MethodDelete)
		api.HandleFunc("/calendar/connections/{id}/sync", rh("TriggerSync", triggerSyncHandler(db, syncTrigger))).Methods(http.MethodPost)
	}
}

// InitializePublicRoutes registers routes that do not require JWT authentication (e.g., OAuth callbacks).
func InitializePublicRoutes(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, cfg config.Config, syncTrigger SyncTrigger) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/calendar/connections/google/callback", rh("GoogleCallback", callbackHandler(db, gcClient, enc, cfg, syncTrigger))).Methods(http.MethodGet)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.ByUserAndHousehold(t.UserId(), t.HouseholdId())
			if err != nil {
				d.Logger().WithError(err).Error("failed to list connections")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}
			rest, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("transforming connections")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func authorizeHandler(db *gorm.DB, gcClient *googlecal.Client, cfg config.Config) server.InputHandler[AuthorizeRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input AuthorizeRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			redirectURI := buildCallbackURI(r)

			stateProc := oauthstate.NewProcessor(d.Logger(), r.Context(), db)
			state, err := stateProc.Create(t.Id(), t.HouseholdId(), t.UserId(), input.RedirectUri, input.Reauthorize)
			if err != nil {
				d.Logger().WithError(err).Error("failed to create oauth state")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}

			forceConsent := input.Reauthorize
			authURL := googlecal.AuthURL(cfg.GoogleCalendarClientID, redirectURI, state.Id().String(), forceConsent)
			d.Logger().WithField("redirect_uri", redirectURI).WithField("auth_url", authURL).Info("generated Google OAuth authorize URL")

			resp := AuthorizeResponse{
				Id:           state.Id(),
				AuthorizeUrl: authURL,
			}
			server.MarshalResponse[AuthorizeResponse](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(resp)
		}
	}
}

func callbackHandler(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, cfg config.Config, syncTrigger SyncTrigger) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			stateParam := r.URL.Query().Get("state")
			code := r.URL.Query().Get("code")

			if stateParam == "" || code == "" {
				http.Redirect(w, r, "/app/calendar?error=missing_params", http.StatusFound)
				return
			}

			stateID, err := uuid.Parse(stateParam)
			if err != nil {
				http.Redirect(w, r, "/app/calendar?error=invalid_state", http.StatusFound)
				return
			}

			stateProc := oauthstate.NewProcessor(d.Logger(), r.Context(), db)
			state, err := stateProc.ValidateAndConsume(stateID)
			if err != nil {
				d.Logger().WithError(err).Warn("OAuth state validation failed")
				http.Redirect(w, r, "/app/calendar?error=invalid_state", http.StatusFound)
				return
			}

			redirectURI := buildCallbackURI(r)
			tokenResp, err := gcClient.ExchangeCode(r.Context(), code, redirectURI)
			if err != nil {
				d.Logger().WithError(err).Error("Google token exchange failed")
				http.Redirect(w, r, "/app/calendar?error=auth_failed", http.StatusFound)
				return
			}

			email, err := gcClient.FetchUserEmail(r.Context(), tokenResp.AccessToken)
			if err != nil || email == "" {
				d.Logger().WithError(err).Warn("failed to fetch Google user email")
				email = "unknown@gmail.com"
			}

			encAccess, err := enc.Encrypt(tokenResp.AccessToken)
			if err != nil {
				d.Logger().WithError(err).Error("failed to encrypt access token")
				http.Redirect(w, r, "/app/calendar?error=internal", http.StatusFound)
				return
			}
			encRefresh, err := enc.Encrypt(tokenResp.RefreshToken)
			if err != nil {
				d.Logger().WithError(err).Error("failed to encrypt refresh token")
				http.Redirect(w, r, "/app/calendar?error=internal", http.StatusFound)
				return
			}

			displayName := email

			tokenExpiry := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

			proc := NewProcessor(d.Logger(), r.Context(), db)

			if state.Reauthorize() {
				existing, findErr := proc.ByUserAndProvider(state.UserID(), "google")
				if findErr != nil {
					d.Logger().WithError(findErr).Error("re-authorization: existing connection not found")
					http.Redirect(w, r, "/app/calendar?error=auth_failed", http.StatusFound)
					return
				}

				if err := proc.UpdateTokensAndWriteAccess(existing.Id(), encAccess, encRefresh, tokenExpiry, true); err != nil {
					d.Logger().WithError(err).Error("failed to update connection for re-authorization")
					http.Redirect(w, r, "/app/calendar?error=internal", http.StatusFound)
					return
				}

				updatedConn, _ := proc.ByIDProvider(existing.Id())()
				if syncTrigger != nil {
					go syncTrigger(updatedConn)
				}

				http.Redirect(w, r, "/app/calendar?connected=true", http.StatusFound)
				return
			}

			conn, err := proc.Create(
				state.TenantID(), state.HouseholdID(), state.UserID(),
				"google", email, encAccess, encRefresh, displayName, tokenExpiry,
			)
			if err != nil {
				d.Logger().WithError(err).Error("failed to create connection")
				http.Redirect(w, r, "/app/calendar?error=already_connected", http.StatusFound)
				return
			}

			if syncTrigger != nil {
				go syncTrigger(conn)
			}

			http.Redirect(w, r, "/app/calendar?connected=true", http.StatusFound)
		}
	}
}

func deleteHandler(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, cascadeDelete CascadeDelete) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)

				conn, err := proc.ByIDProvider(id)()
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Calendar connection not found")
					return
				}
				if conn.UserID() != t.UserId() {
					server.WriteError(w, http.StatusForbidden, "Forbidden", "Connection belongs to another user")
					return
				}

				refreshToken, err := enc.Decrypt(conn.RefreshToken())
				if err == nil {
					_ = gcClient.RevokeToken(r.Context(), refreshToken)
				}

				if cascadeDelete != nil {
					cascadeDelete(r.Context(), id)
				}

				if err := proc.Delete(id); err != nil {
					d.Logger().WithError(err).Error("failed to delete connection")
					server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func triggerSyncHandler(db *gorm.DB, syncTrigger SyncTrigger) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)

				conn, err := proc.ByIDProvider(id)()
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Calendar connection not found")
					return
				}
				if conn.UserID() != t.UserId() {
					server.WriteError(w, http.StatusForbidden, "Forbidden", "Connection belongs to another user")
					return
				}

				if err := proc.CheckManualSyncAllowed(conn); err != nil {
					if errors.Is(err, ErrSyncRateLimited) {
						w.Header().Set("Retry-After", "300")
						server.WriteError(w, http.StatusTooManyRequests, "Rate Limited", "Manual sync allowed once per 5 minutes")
						return
					}
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}

				if syncTrigger != nil {
					go syncTrigger(conn)
				}

				rest, err := Transform(conn)
				if err != nil {
					d.Logger().WithError(err).Error("transforming connection")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func buildCallbackURI(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && !strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "http"
	}
	return scheme + "://" + r.Host + config.CallbackPath
}
