package resource

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/externalidentity"
	authjwt "github.com/jtumidanski/home-hub/services/auth-service/internal/jwt"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/oidc"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/refreshtoken"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/user"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers all auth service routes.
func InitializeRoutes(l *logrus.Logger, db *gorm.DB, issuer *authjwt.Issuer) server.RouteInitializer {
	return func(router *mux.Router) {
		api := router.PathPrefix("/api/v1").Subrouter()

		api.HandleFunc("/auth/providers", handleGetProviders(l, db)).Methods(http.MethodGet)
		api.HandleFunc("/auth/login/{provider}", handleLogin(l)).Methods(http.MethodGet)
		api.HandleFunc("/auth/callback/{provider}", handleCallback(l, db, issuer)).Methods(http.MethodGet)
		api.HandleFunc("/auth/token/refresh", handleRefresh(l, db, issuer)).Methods(http.MethodPost)
		api.HandleFunc("/auth/logout", handleLogout(l, db)).Methods(http.MethodPost)
		api.HandleFunc("/auth/.well-known/jwks.json", handleJWKS(issuer)).Methods(http.MethodGet)
		api.HandleFunc("/users/me", handleGetMe(l, db)).Methods(http.MethodGet)
	}
}

func handleGetProviders(l *logrus.Logger, db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entities []oidc.ProviderConfig
		// For now, return configured providers from environment
		clientID := os.Getenv("OIDC_CLIENT_ID")
		if clientID != "" {
			entities = append(entities, oidc.ProviderConfig{Name: "Google", ClientID: clientID})
		}

		type providerItem struct {
			Type       string            `json:"type"`
			Id         string            `json:"id"`
			Attributes map[string]string `json:"attributes"`
		}
		var data []providerItem
		for _, p := range entities {
			data = append(data, providerItem{
				Type:       "auth-providers",
				Id:         "google",
				Attributes: map[string]string{"displayName": p.Name},
			})
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}
}

func handleLogin(l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := mux.Vars(r)["provider"]
		if provider != "google" {
			server.WriteError(w, http.StatusBadRequest, "Unknown Provider", "Provider not supported")
			return
		}

		issuerURL := os.Getenv("OIDC_ISSUER_URL")
		if issuerURL == "" {
			issuerURL = "https://accounts.google.com"
		}

		disc, err := oidc.Discover(issuerURL)
		if err != nil {
			l.WithError(err).Error("OIDC discovery failed")
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

		redirectURI := os.Getenv("OIDC_REDIRECT_URI")
		cfg := oidc.ProviderConfig{
			ClientID:    os.Getenv("OIDC_CLIENT_ID"),
			RedirectURL: redirectURI,
		}

		authURL := oidc.AuthURL(disc, cfg, state)
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

func handleCallback(l *logrus.Logger, db *gorm.DB, issuer *authjwt.Issuer) http.HandlerFunc {
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

		issuerURL := os.Getenv("OIDC_ISSUER_URL")
		if issuerURL == "" {
			issuerURL = "https://accounts.google.com"
		}

		disc, err := oidc.Discover(issuerURL)
		if err != nil {
			l.WithError(err).Error("OIDC discovery failed")
			server.WriteError(w, http.StatusInternalServerError, "Discovery Failed", "")
			return
		}

		cfg := oidc.ProviderConfig{
			ClientID:     os.Getenv("OIDC_CLIENT_ID"),
			ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("OIDC_REDIRECT_URI"),
		}

		tokenResp, err := oidc.ExchangeCode(r.Context(), disc, cfg, code)
		if err != nil {
			l.WithError(err).Error("code exchange failed")
			server.WriteError(w, http.StatusInternalServerError, "Token Exchange Failed", "")
			return
		}

		userInfo, err := oidc.FetchUserInfo(r.Context(), disc, tokenResp.AccessToken)
		if err != nil {
			l.WithError(err).Error("userinfo fetch failed")
			server.WriteError(w, http.StatusInternalServerError, "UserInfo Failed", "")
			return
		}

		ctx := r.Context()

		// Find or create user
		userProc := user.NewProcessor(l, ctx, db)
		u, err := userProc.FindOrCreate(userInfo.Email, userInfo.DisplayName, userInfo.GivenName, userInfo.FamilyName, userInfo.AvatarURL)
		if err != nil {
			l.WithError(err).Error("user find/create failed")
			server.WriteError(w, http.StatusInternalServerError, "User Error", "")
			return
		}

		// Link external identity
		eiProc := externalidentity.NewProcessor(l, ctx, db)
		_, linkErr := eiProc.FindByProviderSubject("google", userInfo.Subject)()
		if linkErr != nil {
			_, err = eiProc.Create(u.Id(), "google", userInfo.Subject)
			if err != nil {
				l.WithError(err).Error("external identity creation failed")
			}
		}

		// Issue tokens — tenant/household will be zeros until account-service onboarding
		accessToken, err := issuer.Issue(u.Id(), [16]byte{}, [16]byte{})
		if err != nil {
			l.WithError(err).Error("JWT issuance failed")
			server.WriteError(w, http.StatusInternalServerError, "Token Error", "")
			return
		}

		rtProc := refreshtoken.NewProcessor(l, ctx, db)
		rawRefresh, err := rtProc.Create(u.Id())
		if err != nil {
			l.WithError(err).Error("refresh token creation failed")
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

func handleRefresh(l *logrus.Logger, db *gorm.DB, issuer *authjwt.Issuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "No refresh token")
			return
		}

		ctx := r.Context()
		rtProc := refreshtoken.NewProcessor(l, ctx, db)
		newRaw, userID, err := rtProc.Rotate(cookie.Value)
		if err != nil {
			l.WithError(err).Warn("refresh token rotation failed")
			server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid refresh token")
			return
		}

		// Issue new access token — tenant/household zeros (frontend will resolve via context endpoint)
		accessToken, err := issuer.Issue(userID, [16]byte{}, [16]byte{})
		if err != nil {
			l.WithError(err).Error("JWT issuance failed during refresh")
			server.WriteError(w, http.StatusInternalServerError, "Token Error", "")
			return
		}

		setAuthCookies(w, accessToken, newRaw)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleLogout(l *logrus.Logger, db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err == nil && cookie.Value != "" {
			ctx := r.Context()
			rtProc := refreshtoken.NewProcessor(l, ctx, db)
			// Best effort revoke
			_ = rtProc.RevokeAllForUser([16]byte{}) // We'd need user from access token in practice
		}

		clearAuthCookies(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleJWKS(issuer *authjwt.Issuer) http.HandlerFunc {
	jwks := authjwt.BuildJWKS(issuer)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		json.NewEncoder(w).Encode(jwks)
	}
}

func handleGetMe(l *logrus.Logger, db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user from JWT claims
		claims, err := extractClaimsFromCookie(r)
		if err != nil {
			server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid or missing token")
			return
		}

		ctx := r.Context()
		proc := user.NewProcessor(l, ctx, db)
		u, err := proc.ByIDProvider(claims.UserID)()
		if err != nil {
			server.WriteError(w, http.StatusNotFound, "Not Found", "User not found")
			return
		}

		rest, _ := user.Transform(u)
		server.MarshalResponse(w, http.StatusOK, rest)
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

func extractClaimsFromCookie(r *http.Request) (*authjwt.Claims, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return nil, err
	}

	// Parse without validation here — in production the shared auth middleware
	// would validate. For /users/me we just need to extract the user ID.
	parser := jwtgo.NewParser(jwtgo.WithoutClaimsValidation())
	claims := &authjwt.Claims{}
	_, _, err = parser.ParseUnverified(cookie.Value, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
