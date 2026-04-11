package today

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/tz"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB, accountBaseURL string) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/trackers/today", rh("GetTodayTrackers", todayHandler(db, accountBaseURL))).Methods(http.MethodGet)
	}
}

func todayHandler(db *gorm.DB, accountBaseURL string) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			lookup := tz.NewAccountHouseholdLookup(accountBaseURL, r.Header.Get("Authorization"))
			loc := tz.Resolve(r.Context(), d.Logger(), r.Header, t.HouseholdId(), lookup)
			ctx := tz.WithLocation(r.Context(), loc)
			now := time.Now().In(loc)

			proc := NewProcessor(d.Logger(), ctx, db)
			result, err := proc.Today(t.UserId(), now)
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
