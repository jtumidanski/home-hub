package oidcprovider

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/config"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
)

// InitializeRoutes registers OIDC provider routes.
func InitializeRoutes(oidcCfg config.OIDCConfig) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/auth/providers", rh("ListProviders", listHandler(oidcCfg))).Methods(http.MethodGet)
	}
}

func listHandler(oidcCfg config.OIDCConfig) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var providers []RestModel
			if oidcCfg.ClientID != "" {
				providers = append(providers, RestModel{
					Id:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					DisplayName: "Google",
				})
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(providers)
		}
	}
}
