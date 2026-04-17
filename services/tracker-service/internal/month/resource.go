package month

import (
	"errors"
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

		api.HandleFunc("/trackers/months/{month}", rh("GetMonthSummary", monthSummaryHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/trackers/months/{month}/report", rh("GetMonthReport", monthReportHandler(db))).Methods(http.MethodGet)
	}
}

func monthSummaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			today, err := httpparams.ParseDateParam(r, "today")
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
				return
			}

			t := tenantctx.MustFromContext(r.Context())
			monthStr := mux.Vars(r)["month"]

			proc := NewProcessor(d.Logger(), r.Context(), db)
			detail, err := proc.ComputeMonthSummaryDetail(t.UserId(), monthStr, today)
			if err != nil {
				if errors.Is(err, ErrInvalidMonth) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to compute month summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			body, err := MarshalMonthSummary(detail)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to marshal month summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(body); err != nil {
				d.Logger().WithError(err).Warn("Failed to write month summary response")
			}
		}
	}
}

func monthReportHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			today, err := httpparams.ParseDateParam(r, "today")
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
				return
			}

			t := tenantctx.MustFromContext(r.Context())
			monthStr := mux.Vars(r)["month"]

			proc := NewProcessor(d.Logger(), r.Context(), db)
			report, err := proc.ComputeReport(t.UserId(), monthStr, today)
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

			body, err := MarshalReport(report)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to marshal month report")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(body); err != nil {
				d.Logger().WithError(err).Warn("Failed to write month report response")
			}
		}
	}
}

