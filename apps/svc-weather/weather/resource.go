package weather

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/sirupsen/logrus"
)

// HouseholdTracker defines interface for tracking households (passive discovery)
type HouseholdTracker interface {
	TrackHousehold(householdID uuid.UUID)
}

// InitializeRoutes registers weather routes
func InitializeRoutes(si jsonapi.ServerInformation, provider Provider, tracker HouseholdTracker) server.RouteInitializer {
	return func(router *mux.Router, l logrus.FieldLogger) {
		// Public kiosk endpoints
		router.HandleFunc("/weather/combined", server.RegisterHandler(l)(si)("get-combined-weather", getCombinedHandler(provider, tracker))).Methods(http.MethodGet)
		router.HandleFunc("/weather/current", server.RegisterHandler(l)(si)("get-current-weather", getCurrentHandler(provider, tracker))).Methods(http.MethodGet)
		router.HandleFunc("/weather/forecast", server.RegisterHandler(l)(si)("get-forecast-weather", getForecastHandler(provider, tracker))).Methods(http.MethodGet)

		// Admin endpoints
		router.HandleFunc("/admin/weather/cache", server.RegisterHandler(l)(si)("purge-weather-cache", purgeCacheHandler(provider))).Methods(http.MethodDelete)
		router.HandleFunc("/admin/weather/refresh", server.RegisterHandler(l)(si)("refresh-weather-cache", refreshCacheHandler(provider))).Methods(http.MethodPost)
		router.HandleFunc("/admin/weather/cache/all", server.RegisterHandler(l)(si)("purge-all-weather-cache", purgeAllCacheHandler(provider))).Methods(http.MethodDelete)
	}
}

// getCombinedHandler handles GET /weather/combined
func getCombinedHandler(provider Provider, tracker HouseholdTracker) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			householdID, err := parseHouseholdID(d.Logger(), r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Track household for background refresh (passive discovery)
			tracker.TrackHousehold(householdID)

			combined, stale, err := provider.GetCombined(r.Context(), householdID)
			if err != nil {
				d.Logger().WithError(err).WithField("household_id", householdID).Error("Failed to get combined weather")

				// Trigger immediate refresh on cache miss to populate data
				go func() {
					if refreshErr := provider.Refresh(context.Background(), householdID); refreshErr != nil {
						d.Logger().WithError(refreshErr).WithField("household_id", householdID).Warn("Failed to trigger initial refresh")
					} else {
						d.Logger().WithField("household_id", householdID).Info("Triggered initial weather refresh")
					}
				}()

				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			model := Transform(householdID, combined, stale)
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(model)
		}
	}
}

// getCurrentHandler handles GET /weather/current
func getCurrentHandler(provider Provider, tracker HouseholdTracker) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			householdID, err := parseHouseholdID(d.Logger(), r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Track household for background refresh (passive discovery)
			tracker.TrackHousehold(householdID)

			current, stale, err := provider.GetCurrent(r.Context(), householdID)
			if err != nil {
				d.Logger().WithError(err).WithField("household_id", householdID).Error("Failed to get current weather")

				// Trigger immediate refresh on cache miss to populate data
				go func() {
					if refreshErr := provider.RefreshCurrent(context.Background(), householdID); refreshErr != nil {
						d.Logger().WithError(refreshErr).WithField("household_id", householdID).Warn("Failed to trigger initial current refresh")
					} else {
						d.Logger().WithField("household_id", householdID).Info("Triggered initial current weather refresh")
					}
				}()

				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			model := RestModel{
				Id:    householdID,
				Units: "celsius",
				Stale: stale,
				Current: &CurrentAttributes{
					TemperatureC: current.TemperatureC(),
					ObservedAt:   current.ObservedAt().Format("2006-01-02T15:04:05Z07:00"),
					Stale:        stale,
					AgeSeconds:   int64(current.Age().Seconds()),
				},
			}

			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(model)
		}
	}
}

// getForecastHandler handles GET /weather/forecast
func getForecastHandler(provider Provider, tracker HouseholdTracker) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			householdID, err := parseHouseholdID(d.Logger(), r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Track household for background refresh (passive discovery)
			tracker.TrackHousehold(householdID)

			// Parse days parameter (default 7)
			days := 7
			if daysParam := r.URL.Query().Get("days"); daysParam != "" {
				parsed, err := strconv.Atoi(daysParam)
				if err != nil || parsed < 1 || parsed > 14 {
					d.Logger().WithError(err).Error("Invalid days parameter")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				days = parsed
			}

			forecast, stale, err := provider.GetForecast(r.Context(), householdID, days)
			if err != nil {
				d.Logger().WithError(err).WithField("household_id", householdID).Error("Failed to get forecast weather")

				// Trigger immediate refresh on cache miss to populate data
				go func() {
					if refreshErr := provider.RefreshForecast(context.Background(), householdID); refreshErr != nil {
						d.Logger().WithError(refreshErr).WithField("household_id", householdID).Warn("Failed to trigger initial forecast refresh")
					} else {
						d.Logger().WithField("household_id", householdID).Info("Triggered initial forecast weather refresh")
					}
				}()

				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			dailyDays := forecast.Days()
			dailyAttrs := make([]DailyAttributes, len(dailyDays))
			for i, day := range dailyDays {
				dailyAttrs[i] = DailyAttributes{
					Date:  day.Date().Format("2006-01-02"),
					TMaxC: day.TMaxC(),
					TMinC: day.TMinC(),
				}
			}

			model := RestModel{
				Id:    householdID,
				Units: "celsius",
				Stale: stale,
				Daily: dailyAttrs,
			}

			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(model)
		}
	}
}

// purgeCacheHandler handles DELETE /admin/weather/cache
func purgeCacheHandler(provider Provider) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			householdID, err := parseHouseholdID(d.Logger(), r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := provider.Purge(r.Context(), householdID); err != nil {
				d.Logger().WithError(err).WithField("household_id", householdID).Error("Failed to purge cache")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			d.Logger().WithField("household_id", householdID).Info("Cache purged")
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// refreshCacheHandler handles POST /admin/weather/refresh
func refreshCacheHandler(provider Provider) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			householdID, err := parseHouseholdID(d.Logger(), r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Trigger async refresh (non-blocking)
			// Use context.Background() since the HTTP request will complete before the refresh
			go func() {
				if err := provider.Refresh(context.Background(), householdID); err != nil {
					d.Logger().WithError(err).WithField("household_id", householdID).Error("Failed to refresh cache")
				}
			}()

			d.Logger().WithField("household_id", householdID).Info("Cache refresh triggered")
			w.WriteHeader(http.StatusAccepted)
		}
	}
}

// purgeAllCacheHandler handles DELETE /admin/weather/cache/all
func purgeAllCacheHandler(provider Provider) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := provider.PurgeAll(r.Context()); err != nil {
				d.Logger().WithError(err).Error("Failed to purge all cache")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			d.Logger().Info("All cache purged")
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// parseHouseholdID extracts and validates household ID from query parameter
func parseHouseholdID(logger logrus.FieldLogger, r *http.Request) (uuid.UUID, error) {
	householdIDStr := r.URL.Query().Get("householdId")
	if householdIDStr == "" {
		logger.Error("householdId query parameter missing")
		return uuid.Nil, &RequestError{
			Status:  http.StatusBadRequest,
			Message: "householdId query parameter required",
		}
	}

	householdID, err := uuid.Parse(householdIDStr)
	if err != nil {
		logger.WithError(err).Error("Invalid householdId format")
		return uuid.Nil, &RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid householdId format",
		}
	}

	return householdID, nil
}

// RequestError represents a request validation error
type RequestError struct {
	Status  int
	Message string
}

func (e *RequestError) Error() string {
	return e.Message
}
