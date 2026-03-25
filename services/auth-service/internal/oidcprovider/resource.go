package oidcprovider

import (
	"net/http"

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
			proc := NewProcessor(d.Logger(), oidcCfg)
			models, err := proc.ListEnabled()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list OIDC providers")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			restModels, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST models")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(restModels)
		}
	}
}
