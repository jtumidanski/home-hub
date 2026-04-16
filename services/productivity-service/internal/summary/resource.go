package summary

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// resolveTimezone parses the X-Timezone header into a *time.Location. Returns
// time.UTC when the header is missing or invalid.
func resolveTimezone(l logrus.FieldLogger, r *http.Request) *time.Location {
	if hdr := r.Header.Get("X-Timezone"); hdr != "" {
		if loc, err := time.LoadLocation(hdr); err == nil {
			return loc
		}
		l.WithField("header", hdr).Warn("invalid X-Timezone header, falling back to UTC")
	}
	return time.UTC
}

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/summary/tasks", rh("GetTaskSummary", taskSummaryHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/summary/reminders", rh("GetReminderSummary", reminderSummaryHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/summary/dashboard", rh("GetDashboardSummary", dashboardSummaryHandler(db))).Methods(http.MethodGet)
	}
}

func taskSummaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			loc := resolveTimezone(d.Logger(), r)
			now := time.Now().In(loc)
			proc := NewProcessor(d.Logger(), r.Context(), db)
			s, err := proc.TaskSummary(now)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to get task summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalResponse[TaskSummary](d.Logger())(w)(c.ServerInformation())(r.URL.Query())(s)
		}
	}
}

func reminderSummaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := NewProcessor(d.Logger(), r.Context(), db)
			s, err := proc.ReminderSummary()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to get reminder summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalResponse[ReminderSummary](d.Logger())(w)(c.ServerInformation())(r.URL.Query())(s)
		}
	}
}

func dashboardSummaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := NewProcessor(d.Logger(), r.Context(), db)
			s, err := proc.DashboardSummary()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to get dashboard summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalResponse[DashboardSummary](d.Logger())(w)(c.ServerInformation())(r.URL.Query())(s)
		}
	}
}
