package resource

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/config"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/externalidentity"
	authjwt "github.com/jtumidanski/home-hub/services/auth-service/internal/jwt"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/oidc"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/refreshtoken"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/user"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers auth flow routes that orchestrate across domains.
// Domain-specific routes (users/me, providers) are registered in their own packages.
func InitializeRoutes(db *gorm.DB, issuer *authjwt.Issuer, oidcCfg config.OIDCConfig) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/auth/login/{provider}", rh("Login", handleLogin(oidcCfg))).Methods(http.MethodGet)
		api.HandleFunc("/auth/callback/{provider}", rh("Callback", handleCallback(db, issuer, oidcCfg))).Methods(http.MethodGet)
		// Note: refresh and logout are POST endpoints with no request body (tokens come from cookies).
		// RegisterInputHandler requires a JSON:API body, so RegisterHandler is correct here.
		api.HandleFunc("/auth/token/refresh", rh("Refresh", handleRefresh(db, issuer))).Methods(http.MethodPost)
		api.HandleFunc("/auth/logout", rh("Logout", handleLogout(db, issuer))).Methods(http.MethodPost)
		api.HandleFunc("/auth/.well-known/jwks.json", rh("JWKS", handleJWKS(issuer))).Methods(http.MethodGet)
	}
}

func handleLogin(oidcCfg config.OIDCConfig) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			provider := mux.Vars(r)["provider"]
			if provider != "google" {
				server.WriteError(w, http.StatusBadRequest, "Unknown Provider", "Provider not supported")
				return
			}

			disc, err := oidc.Discover(oidcCfg.IssuerURL)
			if err != nil {
				d.Logger().WithError(err).Error("OIDC discovery failed")
				server.WriteError(w, http.StatusInternalServerError, "Discovery Failed", "")
				return
			}

			state := generateState()
			http.SetCookie(w, &http.Cookie{
				Name:     "oauth_state",
				Value:    state,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   300,
			})

			cfg := oidc.ProviderConfig{
				ClientID:    oidcCfg.ClientID,
				RedirectURL: oidcCfg.RedirectURI,
			}

			authURL := oidc.AuthURL(disc, cfg, state)
			http.Redirect(w, r, authURL, http.StatusFound)
		}
	}
}

func handleCallback(db *gorm.DB, issuer *authjwt.Issuer, oidcCfg config.OIDCConfig) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Validate state
			stateCookie, err := r.Cookie("oauth_state")
			if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
				server.WriteError(w, http.StatusBadRequest, "Invalid State", "OAuth state mismatch")
				return
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				server.WriteError(w, http.StatusBadRequest, "Missing Code", "Authorization code required")
				return
			}

			disc, err := oidc.Discover(oidcCfg.IssuerURL)
			if err != nil {
				d.Logger().WithError(err).Error("OIDC discovery failed")
				server.WriteError(w, http.StatusInternalServerError, "Discovery Failed", "")
				return
			}

			cfg := oidc.ProviderConfig{
				ClientID:     oidcCfg.ClientID,
				ClientSecret: oidcCfg.ClientSecret,
				RedirectURL:  oidcCfg.RedirectURI,
			}

			tokenResp, err := oidc.ExchangeCode(r.Context(), disc, cfg, code)
			if err != nil {
				d.Logger().WithError(err).Error("code exchange failed")
				server.WriteError(w, http.StatusInternalServerError, "Token Exchange Failed", "")
				return
			}

			userInfo, err := oidc.FetchUserInfo(r.Context(), disc, tokenResp.AccessToken)
			if err != nil {
				d.Logger().WithError(err).Error("userinfo fetch failed")
				server.WriteError(w, http.StatusInternalServerError, "UserInfo Failed", "")
				return
			}

			ctx := r.Context()

			// Find or create user
			userProc := user.NewProcessor(d.Logger(), ctx, db)
			u, err := userProc.FindOrCreate(userInfo.Email, userInfo.DisplayName, userInfo.GivenName, userInfo.FamilyName, userInfo.AvatarURL)
			if err != nil {
				d.Logger().WithError(err).Error("user find/create failed")
				server.WriteError(w, http.StatusInternalServerError, "User Error", "")
				return
			}

			// Link external identity (idempotent — skip if already linked)
			eiProc := externalidentity.NewProcessor(d.Logger(), ctx, db)
			_, linkErr := eiProc.FindByProviderSubject("google", userInfo.Subject)()
			if linkErr != nil {
				_, err = eiProc.Create(u.Id(), "google", userInfo.Subject)
				if err != nil {
					d.Logger().WithError(err).Error("external identity creation failed")
					server.WriteError(w, http.StatusInternalServerError, "Identity Error", "")
					return
				}
			}

			// Issue tokens — tenant/household will be zeros until account-service onboarding
			accessToken, err := issuer.Issue(u.Id(), [16]byte{}, [16]byte{})
			if err != nil {
				d.Logger().WithError(err).Error("JWT issuance failed")
				server.WriteError(w, http.StatusInternalServerError, "Token Error", "")
				return
			}

			rtProc := refreshtoken.NewProcessor(d.Logger(), ctx, db)
			rawRefresh, err := rtProc.Create(u.Id())
			if err != nil {
				d.Logger().WithError(err).Error("refresh token creation failed")
				server.WriteError(w, http.StatusInternalServerError, "Token Error", "")
				return
			}

			setAuthCookies(w, accessToken, rawRefresh)

			// Clear state cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "oauth_state",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})

			redirect := r.URL.Query().Get("redirect")
			if redirect == "" {
				redirect = "/app"
			}
			http.Redirect(w, r, redirect, http.StatusFound)
		}
	}
}

func handleRefresh(db *gorm.DB, issuer *authjwt.Issuer) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("refresh_token")
			if err != nil {
				server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "No refresh token")
				return
			}

			ctx := r.Context()
			rtProc := refreshtoken.NewProcessor(d.Logger(), ctx, db)
			newRaw, userID, err := rtProc.Rotate(cookie.Value)
			if err != nil {
				d.Logger().WithError(err).Warn("refresh token rotation failed")
				server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid refresh token")
				return
			}

			// Issue new access token — tenant/household zeros (frontend will resolve via context endpoint)
			accessToken, err := issuer.Issue(userID, [16]byte{}, [16]byte{})
			if err != nil {
				d.Logger().WithError(err).Error("JWT issuance failed during refresh")
				server.WriteError(w, http.StatusInternalServerError, "Token Error", "")
				return
			}

			setAuthCookies(w, accessToken, newRaw)
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleLogout(db *gorm.DB, issuer *authjwt.Issuer) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract user from access token to revoke their refresh tokens
			claims, err := authjwt.ExtractClaimsFromCookie(r, issuer.PublicKey())
			if err == nil {
				ctx := r.Context()
				rtProc := refreshtoken.NewProcessor(d.Logger(), ctx, db)
				if revokeErr := rtProc.RevokeAllForUser(claims.UserID); revokeErr != nil {
					d.Logger().WithError(revokeErr).Warn("failed to revoke refresh tokens during logout")
				}
			}

			clearAuthCookies(w)
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleJWKS(issuer *authjwt.Issuer) server.GetHandler {
	jwks := authjwt.BuildJWKS(issuer)
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			json.NewEncoder(w).Encode(jwks)
		}
	}
}

func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   900, // 15 minutes
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   604800, // 7 days
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
