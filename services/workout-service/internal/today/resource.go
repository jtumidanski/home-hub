// Package today serves the GET /workouts/today endpoint — the mobile-default
// landing view that returns the current day's planned items + embedded
// performances.
//
// The projection itself lives in processor.go; this file only registers the
// route and writes the JSON envelope.
package today

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/tz"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/weekview"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB, accountBaseURL string) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/workouts/today", rh("GetWorkoutsToday", todayHandler(db, accountBaseURL))).Methods(http.MethodGet)
	}
}

// RestModel is the JSON:API resource for the today projection. The id is
// today's ISO date so the URL identifier and the resource identifier match.
type RestModel struct {
	Id            string              `json:"-"`
	Date          string              `json:"date"`
	WeekStartDate string              `json:"weekStartDate"`
	DayOfWeek     int                 `json:"dayOfWeek"`
	IsRestDay     bool                `json:"isRestDay"`
	Items         []weekview.ItemRest `json:"items"`
}

func (r RestModel) GetName() string         { return "today" }
func (r RestModel) GetID() string           { return r.Id }
func (r *RestModel) SetID(id string) error { r.Id = id; return nil }

func transform(res Result) RestModel {
	items := res.Items
	if items == nil {
		items = []weekview.ItemRest{}
	}
	d := res.Date.Format("2006-01-02")
	return RestModel{
		Id:            d,
		Date:          d,
		WeekStartDate: res.WeekStartDate.Format("2006-01-02"),
		DayOfWeek:     res.DayOfWeek,
		IsRestDay:     res.IsRestDay,
		Items:         items,
	}
}

func todayHandler(db *gorm.DB, accountBaseURL string) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			lookup := tz.NewAccountHouseholdLookup(accountBaseURL, r.Header.Get("Authorization"))
			loc := tz.Resolve(r.Context(), d.Logger(), r.Header, t.HouseholdId(), lookup)
			ctx := tz.WithLocation(r.Context(), loc)
			res, err := NewProcessor(d.Logger(), ctx, db).Today(t.UserId(), time.Now().In(loc))
			if err != nil {
				d.Logger().WithError(err).Error("Failed to load today")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(transform(res))
		}
	}
}
