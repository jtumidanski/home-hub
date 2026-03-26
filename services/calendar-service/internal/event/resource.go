package event

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/calendar/events", rh("ListEvents", listEventsHandler(db))).Methods(http.MethodGet)
	}
}

func listEventsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			now := time.Now().UTC()
			startStr := r.URL.Query().Get("start")
			endStr := r.URL.Query().Get("end")

			var start, end time.Time
			var err error

			if startStr != "" {
				start, err = time.Parse(time.RFC3339, startStr)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Parameter", "Invalid start date format, use ISO 8601")
					return
				}
			} else {
				start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			}

			if endStr != "" {
				end, err = time.Parse(time.RFC3339, endStr)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Parameter", "Invalid end date format, use ISO 8601")
					return
				}
			} else {
				end = start.AddDate(0, 0, 7)
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.QueryByHouseholdAndTimeRange(t.HouseholdId(), start, end)
			if err != nil {
				if errors.Is(err, ErrRangeTooLarge) {
					server.WriteError(w, http.StatusBadRequest, "Range Too Large", "Maximum query range is 90 days")
					return
				}
				d.Logger().WithError(err).Error("failed to query events")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}

			rest, err := TransformSliceWithPrivacy(models, t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("transforming events")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}
