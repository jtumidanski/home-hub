// Package today serves the GET /workouts/today endpoint — the mobile-default
// landing view that returns the current day's planned items + embedded
// performances.
//
// The "current day" is supplied by the client as a `?date=YYYY-MM-DD` query
// parameter. The server does not resolve a timezone — the client knows its
// own local date.
package today

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	httpparams "github.com/jtumidanski/home-hub/shared/go/http"
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
func (r *RestModel) SetID(id string) error  { r.Id = id; return nil }

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

func todayHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			date, err := httpparams.ParseDateParam(r, "date")
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
				return
			}

			t := tenantctx.MustFromContext(r.Context())
			res, err := NewProcessor(d.Logger(), r.Context(), db).Today(t.UserId(), date)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to load today")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(transform(res))
		}
	}
}
