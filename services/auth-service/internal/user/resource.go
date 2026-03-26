package user

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	authjwt "github.com/jtumidanski/home-hub/services/auth-service/internal/jwt"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers user domain routes.
func InitializeRoutes(db *gorm.DB, issuer *authjwt.Issuer) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rih := server.RegisterInputHandler[UpdateRequest](l)(si)
		api.HandleFunc("/users/me", rh("GetMe", getMeHandler(db, issuer))).Methods(http.MethodGet)
		api.HandleFunc("/users/me", rih("PatchMe", patchMeHandler(db, issuer))).Methods(http.MethodPatch)
		api.HandleFunc("/users", rh("ListUsers", listUsersHandler(db, issuer))).Methods(http.MethodGet)
	}
}

func getMeHandler(db *gorm.DB, issuer *authjwt.Issuer) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, err := authjwt.ExtractClaimsFromCookie(r, issuer.PublicKey())
			if err != nil {
				server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid or missing token")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			u, err := proc.ByIDProvider(claims.UserID)()
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "User not found")
				return
			}

			rest, err := Transform(u)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func patchMeHandler(db *gorm.DB, issuer *authjwt.Issuer) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, err := authjwt.ExtractClaimsFromCookie(r, issuer.PublicKey())
			if err != nil {
				server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid or missing token")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			updated, err := proc.UpdateAvatar(claims.UserID, input.AvatarURL)
			if err != nil {
				if errors.Is(err, ErrInvalidAvatarFormat) {
					server.WriteError(w, http.StatusUnprocessableEntity, "Invalid Avatar", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Updating avatar")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			rest, err := Transform(updated)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

const maxBatchIDs = 50

func listUsersHandler(db *gorm.DB, issuer *authjwt.Issuer) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Require authentication
			_, err := authjwt.ExtractClaimsFromCookie(r, issuer.PublicKey())
			if err != nil {
				server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid or missing token")
				return
			}

			idsParam := r.URL.Query().Get("filter[ids]")
			if idsParam == "" {
				server.WriteError(w, http.StatusBadRequest, "Missing Filter", "filter[ids] is required")
				return
			}

			idStrs := strings.Split(idsParam, ",")
			if len(idStrs) > maxBatchIDs {
				server.WriteError(w, http.StatusBadRequest, "Too Many IDs", "maximum 50 IDs per request")
				return
			}

			ids := make([]uuid.UUID, 0, len(idStrs))
			for _, s := range idStrs {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				id, err := uuid.Parse(s)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid ID", "each ID must be a valid UUID")
					return
				}
				ids = append(ids, id)
			}

			if len(ids) == 0 {
				server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())([]RestModel{})
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.ByIDsProvider(ids)()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list users by IDs")
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
