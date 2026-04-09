// Package today serves the GET /workouts/today endpoint — the mobile-default
// landing view that returns the current day's planned items + embedded
// performances. The current day is computed in UTC, matching the existing
// tracker-service today behavior. (PRD §6 calls for the user's TZ; we keep
// parity with tracker-service rather than reinventing TZ resolution here. If
// account-service grows a shared TZ helper later, both today endpoints can
// adopt it together.)
package today

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
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
			now := time.Now().UTC().Truncate(24 * time.Hour)
			weekStart := week.NormalizeToMonday(now)
			// ISO day-of-week with Monday = 0.
			dayOfWeek := (int(now.Weekday()) + 6) % 7

			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			m, err := weekProc.Get(t.UserId(), weekStart)
			if err != nil {
				if errors.Is(err, week.ErrNotFound) {
					writeEmpty(w, now, weekStart, dayOfWeek, false)
					return
				}
				d.Logger().WithError(err).Error("Failed to load today week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			items, err := weekview.AssembleItems(weekProc.DB(), m.Id())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to assemble today items")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			// Filter to today's day-of-week.
			filtered := make([]weekview.ItemRest, 0, len(items))
			for _, it := range items {
				if it.DayOfWeek == dayOfWeek {
					filtered = append(filtered, it)
				}
			}

			isRestDay := false
			for _, d := range m.RestDayFlags() {
				if d == dayOfWeek {
					isRestDay = true
					break
				}
			}

			writeDoc(w, now, weekStart, dayOfWeek, isRestDay, filtered)
		}
	}
}

func writeEmpty(w http.ResponseWriter, today, weekStart time.Time, dayOfWeek int, isRestDay bool) {
	writeDoc(w, today, weekStart, dayOfWeek, isRestDay, []weekview.ItemRest{})
}

func writeDoc(w http.ResponseWriter, today, weekStart time.Time, dayOfWeek int, isRestDay bool, items []weekview.ItemRest) {
	doc := document{
		Data: data{
			Type: "today",
			ID:   today.Format("2006-01-02"),
			Attributes: attributes{
				Date:          today.Format("2006-01-02"),
				WeekStartDate: weekStart.Format("2006-01-02"),
				DayOfWeek:     dayOfWeek,
				IsRestDay:     isRestDay,
				Items:         items,
			},
		},
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(doc)
}
