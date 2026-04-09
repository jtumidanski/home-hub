// Package today serves the GET /workouts/today endpoint — the mobile-default
// landing view that returns the current day's planned items + embedded
// performances.
//
// The projection itself lives in processor.go; this file only registers the
// route and writes the JSON envelope.
package today

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/weekview"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/workouts/today", rh("GetWorkoutsToday", todayHandler(db))).Methods(http.MethodGet)
	}
}

type document struct {
	Data data `json:"data"`
}

type data struct {
	Type       string     `json:"type"`
	ID         string     `json:"id"`
	Attributes attributes `json:"attributes"`
}

type attributes struct {
	Date          string              `json:"date"`
	WeekStartDate string              `json:"weekStartDate"`
	DayOfWeek     int                 `json:"dayOfWeek"`
	IsRestDay     bool                `json:"isRestDay"`
	Items         []weekview.ItemRest `json:"items"`
}

func todayHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			res, err := NewProcessor(d.Logger(), r.Context(), db).Today(t.UserId(), time.Now())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to load today")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			writeDoc(w, d.Logger(), res)
		}
	}
}

func writeDoc(w http.ResponseWriter, l logrus.FieldLogger, res Result) {
	doc := document{
		Data: data{
			Type: "today",
			ID:   res.Date.Format("2006-01-02"),
			Attributes: attributes{
				Date:          res.Date.Format("2006-01-02"),
				WeekStartDate: res.WeekStartDate.Format("2006-01-02"),
				DayOfWeek:     res.DayOfWeek,
				IsRestDay:     res.IsRestDay,
				Items:         res.Items,
			},
		},
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		l.WithError(err).Error("Failed to encode today document")
	}
}
