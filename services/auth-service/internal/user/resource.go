package user

import (
	"net/http"

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
		api.HandleFunc("/users/me", rh("GetMe", getMeHandler(db, issuer))).Methods(http.MethodGet)
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
