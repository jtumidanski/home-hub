// Package summary serves the GET /workouts/weeks/{weekStart}/summary endpoint
// — the per-week reporting projection that powers the weekly summary screen.
//
// The projection itself lives in processor.go; this file only registers the
// route, parses the URL parameter, and writes the JSON envelope.
package summary

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const weekDateLayout = "2006-01-02"

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/workouts/weeks/{weekStart}/summary", rh("GetWeekSummary", summaryHandler(db))).Methods(http.MethodGet)
	}
}

func summaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			ws, err := time.ParseInLocation(weekDateLayout, mux.Vars(r)["weekStart"], time.UTC)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}

			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			wkModel, err := weekProc.Get(t.UserId(), ws)
			if err != nil {
				if errors.Is(err, week.ErrNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Week not found")
					return
				}
				d.Logger().WithError(err).Error("Failed to load week for summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			doc, err := NewProcessor(d.Logger(), r.Context(), db).Build(wkModel)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to build week summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(doc); err != nil {
				d.Logger().WithError(err).Error("Failed to encode summary document")
			}
		}
	}
}
