package month

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/trackers/months/{month}", rh("GetMonthSummary", monthSummaryHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/trackers/months/{month}/report", rh("GetMonthReport", monthReportHandler(db))).Methods(http.MethodGet)
	}
}

func monthSummaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			monthStr := mux.Vars(r)["month"]

			proc := NewProcessor(d.Logger(), r.Context(), db)
			summary, items, entries, err := proc.ComputeMonthSummary(t.UserId(), monthStr)
			if err != nil {
				if errors.Is(err, ErrInvalidMonth) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to compute month summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			itemIDs := make([]uuid.UUID, len(items))
			for i, m := range items {
				itemIDs[i] = m.Id()
			}

			snapshotsByItem := make(map[uuid.UUID][]schedule.Model)
			if len(itemIDs) > 0 {
				snapEntities, _ := schedule.GetByTrackingItemIDs(itemIDs)(db.WithContext(r.Context()))()
				for _, se := range snapEntities {
					sm, err := schedule.Make(se)
					if err != nil {
						continue
					}
					snapshotsByItem[sm.TrackingItemID()] = append(snapshotsByItem[sm.TrackingItemID()], sm)
				}
			}

			entryModels := make([]entry.Model, 0)
			for _, e := range entries {
				entryModels = append(entryModels, e)
			}

			itemModels := make([]trackingitem.Model, 0)
			for _, m := range items {
				itemModels = append(itemModels, m)
			}

			result := TransformMonthSummary(summary, itemModels, entryModels, snapshotsByItem)

			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			w.Write(result)
		}
	}
}

func monthReportHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			monthStr := mux.Vars(r)["month"]

			proc := NewProcessor(d.Logger(), r.Context(), db)
			report, err := proc.ComputeReport(t.UserId(), monthStr)
			if err != nil {
				if errors.Is(err, ErrInvalidMonth) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				if errors.Is(err, ErrMonthIncomplete) {
					server.WriteError(w, http.StatusBadRequest, "Month Incomplete", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to compute month report")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest := struct {
				Data struct {
					Type       string `json:"type"`
					Attributes Report `json:"attributes"`
				} `json:"data"`
			}{}
			rest.Data.Type = "tracker-reports"
			rest.Data.Attributes = report

			b, _ := json.Marshal(rest)
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		}
	}
}
