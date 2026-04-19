package weekview

import (
	"errors"
	"net/http"
	"time"

	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"gorm.io/gorm"
)

// NearestPointer is the JSON:API resource returned by GET /weeks/nearest. The
// `id` is the week's ISO start date so the URL identifier and the resource
// identifier match the pattern used by the week-summary endpoint.
type NearestPointer struct {
	Id            string `json:"-"`
	WeekStartDate string `json:"weekStartDate"`
}

func (r NearestPointer) GetName() string         { return "workoutWeekPointer" }
func (r NearestPointer) GetID() string           { return r.Id }
func (r *NearestPointer) SetID(id string) error { r.Id = id; return nil }

// normalizeToMonday rolls a calendar date back to the Monday of the same
// ISO week. Mirrors the normalization the `/weeks/{weekStart}` endpoints
// apply in the week processor.
func normalizeToMonday(d time.Time) time.Time {
	iso := (int(d.Weekday()) + 6) % 7
	return d.AddDate(0, 0, -iso)
}

func nearestHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			ref := r.URL.Query().Get("reference")
			if ref == "" {
				server.WriteError(w, http.StatusBadRequest, "Invalid reference", "reference is required")
				return
			}
			parsed, err := time.ParseInLocation(weekDateLayout, ref, time.UTC)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid reference", "reference must be YYYY-MM-DD")
				return
			}

			direction := r.URL.Query().Get("direction")
			if direction != "prev" && direction != "next" {
				server.WriteError(w, http.StatusBadRequest, "Invalid direction", "direction must be 'prev' or 'next'")
				return
			}

			monday := normalizeToMonday(parsed)

			var entity week.Entity
			scopedDB := db.WithContext(r.Context())
			switch direction {
			case "prev":
				entity, err = week.GetMostRecentPriorWithItems(scopedDB, t.UserId(), monday)
			case "next":
				entity, err = week.GetSoonestNextWithItems(scopedDB, t.UserId(), monday)
			}
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "No populated week in that direction")
					return
				}
				d.Logger().WithError(err).Error("Failed to load nearest populated week")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			wkDate := entity.WeekStartDate.Format("2006-01-02")
			pointer := NearestPointer{Id: wkDate, WeekStartDate: wkDate}
			server.MarshalResponse[NearestPointer](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(pointer)
		}
	}
}
