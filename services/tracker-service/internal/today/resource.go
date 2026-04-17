package today

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	httpparams "github.com/jtumidanski/home-hub/shared/go/http"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/trackers/today", rh("GetTodayTrackers", todayHandler(db))).Methods(http.MethodGet)
	}
}

func todayHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			date, err := httpparams.ParseDateParam(r, "date")
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
				return
			}

			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			result, err := proc.Today(t.UserId(), date)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to compute today view")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			body, err := MarshalDocument(Transform(result))
			if err != nil {
				d.Logger().WithError(err).Error("Failed to marshal today document")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(body); err != nil {
				d.Logger().WithError(err).Warn("Failed to write today response")
			}
		}
	}
}
