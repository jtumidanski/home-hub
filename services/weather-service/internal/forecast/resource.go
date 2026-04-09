package forecast

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/locationofinterest"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB, client *openmeteo.Client, cacheTTL time.Duration) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/weather/current", rh("GetCurrentWeather", currentHandler(db, client, cacheTTL))).Methods(http.MethodGet)
		api.HandleFunc("/weather/forecast", rh("GetWeatherForecast", forecastHandler(db, client, cacheTTL))).Methods(http.MethodGet)
	}
}

// resolveLocation resolves the optional locationId query parameter. Returns
// (locationID, lat, lon, ok). When ok is false the response has been written.
func resolveLocation(w http.ResponseWriter, r *http.Request, db *gorm.DB, d *server.HandlerDependency) (*uuid.UUID, float64, float64, bool) {
	t := tenantctx.MustFromContext(r.Context())

	locStr := r.URL.Query().Get("locationId")
	if locStr != "" {
		id, err := uuid.Parse(locStr)
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "Invalid locationId", "locationId must be a valid UUID")
			return nil, 0, 0, false
		}
		proc := locationofinterest.NewProcessor(d.Logger(), r.Context(), db, nil)
		m, err := proc.Get(t.HouseholdId(), id)
		if err != nil {
			if errors.Is(err, locationofinterest.ErrNotFound) {
				server.WriteError(w, http.StatusNotFound, "Not Found", "Location of interest not found")
				return nil, 0, 0, false
			}
			d.Logger().WithError(err).Error("Failed to resolve locationId")
			server.WriteError(w, http.StatusInternalServerError, "Error", "")
			return nil, 0, 0, false
		}
		idCopy := id
		return &idCopy, m.Latitude(), m.Longitude(), true
	}

	lat := r.URL.Query().Get("latitude")
	lon := r.URL.Query().Get("longitude")
	if lat == "" || lon == "" {
		server.WriteError(w, http.StatusNotFound, "No Location", "The active household does not have a location configured. Set a location in household settings.")
		return nil, 0, 0, false
	}
	latF, lonF, err := parseCoordinates(lat, lon)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, "Invalid Coordinates", err.Error())
		return nil, 0, 0, false
	}
	return nil, latF, lonF, true
}

func currentHandler(db *gorm.DB, client *openmeteo.Client, cacheTTL time.Duration) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			locationID, latF, lonF, ok := resolveLocation(w, r, db, d)
			if !ok {
				return
			}

			units := r.URL.Query().Get("units")
			timezone := r.URL.Query().Get("timezone")
			if units == "" {
				units = "metric"
			}
			if timezone == "" {
				timezone = "UTC"
			}

			proc := NewProcessor(d.Logger(), r.Context(), db, client, cacheTTL)
			m, err := proc.GetCurrent(t.Id(), t.HouseholdId(), locationID, latF, lonF, units, timezone)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to get current weather")
				server.WriteError(w, http.StatusBadGateway, "Weather Unavailable", "Unable to retrieve weather data. Please try again later.")
				return
			}

			rest, err := TransformCurrent(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalResponse[CurrentRestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func forecastHandler(db *gorm.DB, client *openmeteo.Client, cacheTTL time.Duration) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			locationID, latF, lonF, ok := resolveLocation(w, r, db, d)
			if !ok {
				return
			}

			units := r.URL.Query().Get("units")
			timezone := r.URL.Query().Get("timezone")
			if units == "" {
				units = "metric"
			}
			if timezone == "" {
				timezone = "UTC"
			}

			proc := NewProcessor(d.Logger(), r.Context(), db, client, cacheTTL)
			m, err := proc.GetForecast(t.Id(), t.HouseholdId(), locationID, latF, lonF, units, timezone)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to get weather forecast")
				server.WriteError(w, http.StatusBadGateway, "Weather Unavailable", "Unable to retrieve weather data. Please try again later.")
				return
			}

			rest, err := TransformForecast(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalSliceResponse[DailyRestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func parseCoordinates(lat, lon string) (float64, float64, error) {
	var latF, lonF float64
	_, err := parseFloat(lat, &latF)
	if err != nil {
		return 0, 0, err
	}
	_, err = parseFloat(lon, &lonF)
	if err != nil {
		return 0, 0, err
	}
	return latF, lonF, nil
}

func parseFloat(s string, f *float64) (float64, error) {
	var err error
	_, err = fmt.Sscanf(s, "%f", f)
	return *f, err
}
