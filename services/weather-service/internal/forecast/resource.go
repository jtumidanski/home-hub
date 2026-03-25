package forecast

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB, client *openmeteo.Client) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/weather/current", rh("GetCurrentWeather", currentHandler(db, client))).Methods(http.MethodGet)
		api.HandleFunc("/weather/forecast", rh("GetWeatherForecast", forecastHandler(db, client))).Methods(http.MethodGet)
	}
}

func currentHandler(db *gorm.DB, client *openmeteo.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			lat := r.URL.Query().Get("latitude")
			lon := r.URL.Query().Get("longitude")
			units := r.URL.Query().Get("units")
			timezone := r.URL.Query().Get("timezone")

			if lat == "" || lon == "" {
				server.WriteError(w, http.StatusNotFound, "No Location", "The active household does not have a location configured. Set a location in household settings.")
				return
			}

			latF, lonF, err := parseCoordinates(lat, lon)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Coordinates", err.Error())
				return
			}

			if units == "" {
				units = "metric"
			}
			if timezone == "" {
				timezone = "UTC"
			}

			proc := NewProcessor(d.Logger(), r.Context(), db, client)
			m, err := proc.GetCurrent(t.Id(), t.HouseholdId(), latF, lonF, units, timezone)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to get current weather")
				server.WriteError(w, http.StatusBadGateway, "Weather Unavailable", "Unable to retrieve weather data. Please try again later.")
				return
			}

			rest := TransformCurrent(m)
			server.MarshalResponse[CurrentRestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func forecastHandler(db *gorm.DB, client *openmeteo.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			lat := r.URL.Query().Get("latitude")
			lon := r.URL.Query().Get("longitude")
			units := r.URL.Query().Get("units")
			timezone := r.URL.Query().Get("timezone")

			if lat == "" || lon == "" {
				server.WriteError(w, http.StatusNotFound, "No Location", "The active household does not have a location configured. Set a location in household settings.")
				return
			}

			latF, lonF, err := parseCoordinates(lat, lon)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Coordinates", err.Error())
				return
			}

			if units == "" {
				units = "metric"
			}
			if timezone == "" {
				timezone = "UTC"
			}

			proc := NewProcessor(d.Logger(), r.Context(), db, client)
			m, err := proc.GetForecast(t.Id(), t.HouseholdId(), latF, lonF, units, timezone)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to get weather forecast")
				server.WriteError(w, http.StatusBadGateway, "Weather Unavailable", "Unable to retrieve weather data. Please try again later.")
				return
			}

			rest := TransformForecast(m)
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
