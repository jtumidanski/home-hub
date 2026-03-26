package appcontext

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/contexts/current", rh("GetContext", getHandler(db))).Methods(http.MethodGet)
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			var userEmail string
			if claims, ok := sharedauth.ClaimsFromContext(r.Context()); ok {
				userEmail = claims.Email
			}
			resolved, err := Resolve(d.Logger(), r.Context(), db, t.Id(), t.UserId(), userEmail)
			if err != nil {
				d.Logger().WithError(err).Error("Context resolution failed")
				server.WriteError(w, http.StatusInternalServerError, "Context Resolution Failed", err.Error())
				return
			}

			rest, err := TransformContext(resolved)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(r.URL.Query())(rest)
		}
	}
}
