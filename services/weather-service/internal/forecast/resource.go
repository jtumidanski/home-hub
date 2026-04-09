package forecast

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MakeResolver is a factory the application wires at route-init time so each
// request gets a context-bound LocationResolver. Defined as a parameter so the
// forecast package never imports locationofinterest directly.
type MakeResolver func(l logrus.FieldLogger, r *http.Request) LocationResolver

func InitializeRoutes(db *gorm.DB, client *openmeteo.Client, cacheTTL time.Duration, makeResolver MakeResolver) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/weather/current", rh("GetCurrentWeather", currentHandler(db, client, cacheTTL, makeResolver))).Methods(http.MethodGet)
		api.HandleFunc("/weather/forecast", rh("GetWeatherForecast", forecastHandler(db, client, cacheTTL, makeResolver))).Methods(http.MethodGet)
	}
}

// parseLocationRequest parses query params into a LocationRequest. Returns
// (req, ok). When ok is false an error response has already been written.
func parseLocationRequest(w http.ResponseWriter, r *http.Request) (LocationRequest, bool) {
	locStr := r.URL.Query().Get("locationId")
	if locStr != "" {
		id, err := uuid.Parse(locStr)
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "Invalid locationId", "locationId must be a valid UUID")
			return LocationRequest{}, false
		}
		return LocationRequest{LocationID: &id}, true
	}

	lat := r.URL.Query().Get("latitude")
	lon := r.URL.Query().Get("longitude")
	if lat == "" || lon == "" {
		server.WriteError(w, http.StatusNotFound, "No Location", "The active household does not have a location configured. Set a location in household settings.")
		return LocationRequest{}, false
	}
	latF, lonF, err := parseCoordinates(lat, lon)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, "Invalid Coordinates", err.Error())
		return LocationRequest{}, false
	}
	return LocationRequest{Latitude: latF, Longitude: lonF, HasCoords: true}, true
}

func unitsAndTimezone(r *http.Request) (string, string) {
	units := r.URL.Query().Get("units")
	timezone := r.URL.Query().Get("timezone")
	if units == "" {
		units = "metric"
	}
	if timezone == "" {
		timezone = "UTC"
	}
	return units, timezone
}

func currentHandler(db *gorm.DB, client *openmeteo.Client, cacheTTL time.Duration, makeResolver MakeResolver) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			req, ok := parseLocationRequest(w, r)
			if !ok {
				return
			}
			units, timezone := unitsAndTimezone(r)

			var resolver LocationResolver
			if makeResolver != nil {
				resolver = makeResolver(d.Logger(), r)
			}
			proc := NewProcessor(d.Logger(), r.Context(), db, client, cacheTTL, resolver)
			m, err := proc.GetCurrent(t.Id(), t.HouseholdId(), req, units, timezone)
			if err != nil {
				if errors.Is(err, ErrLocationNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Location of interest not found")
					return
				}
				if errors.Is(err, ErrNoLocation) {
					server.WriteError(w, http.StatusNotFound, "No Location", "The active household does not have a location configured.")
					return
				}
				d.Logger().WithError(err).Error("Failed to get current weather")
				server.WriteError(w, http.StatusBadGateway, "Weather Unavailable", "Unable to retrieve weather data. Please try again later.")
				return
			}

			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func forecastHandler(db *gorm.DB, client *openmeteo.Client, cacheTTL time.Duration, makeResolver MakeResolver) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			req, ok := parseLocationRequest(w, r)
			if !ok {
				return
			}
			units, timezone := unitsAndTimezone(r)

			var resolver LocationResolver
			if makeResolver != nil {
				resolver = makeResolver(d.Logger(), r)
			}
			proc := NewProcessor(d.Logger(), r.Context(), db, client, cacheTTL, resolver)
			m, err := proc.GetForecast(t.Id(), t.HouseholdId(), req, units, timezone)
			if err != nil {
				if errors.Is(err, ErrLocationNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Location of interest not found")
					return
				}
				if errors.Is(err, ErrNoLocation) {
					server.WriteError(w, http.StatusNotFound, "No Location", "The active household does not have a location configured.")
					return
				}
				d.Logger().WithError(err).Error("Failed to get weather forecast")
				server.WriteError(w, http.StatusBadGateway, "Weather Unavailable", "Unable to retrieve weather data. Please try again later.")
				return
			}

			rest, err := TransformDaily(m)
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
