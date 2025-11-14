package preference

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers all device preference-related routes
func InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Device preferences endpoints
			router.HandleFunc("/devices/{id}/preferences", server.RegisterHandler(l)(si)("get-device-preferences", getDevicePreferencesHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/devices/{id}/preferences", server.RegisterInputHandler[UpdatePreferencesRequest](l)(si)("update-device-preferences", updateDevicePreferencesHandler(db))).Methods(http.MethodPut)
		}
	}
}

// IdHandler a handler interface which requires a deviceId
type IdHandler func(deviceId uuid.UUID) http.HandlerFunc

// ParseId parses the deviceId consistently from the request, and provide it to the next handler
func ParseId(l logrus.FieldLogger, next IdHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			l.WithError(err).Errorf("Unable to properly parse id from path.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next(id)(w, r)
	}
}

// getDevicePreferencesHandler handles GET /devices/:id/preferences
func getDevicePreferencesHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(deviceId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				model, err := NewProcessor(d.Logger(), r.Context(), db).GetByDeviceId(deviceId)()
				if err != nil {
					if errors.Is(err, ErrPreferencesNotFound) {
						d.Logger().WithError(err).Error("Device preferences not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to fetch device preferences")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				res, err := ops.Map(Transform)(ops.FixedProvider(model))()
				if err != nil {
					d.Logger().WithError(err).Errorf("Creating REST model.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
			}
		})
	}
}

// updateDevicePreferencesHandler handles PUT /devices/:id/preferences
func updateDevicePreferencesHandler(db *gorm.DB) server.InputHandler[UpdatePreferencesRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req UpdatePreferencesRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(deviceId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				input := UpdateInput{
					Theme:           req.Theme,
					TemperatureUnit: req.TemperatureUnit,
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).Upsert(deviceId, input)()
				if err != nil {
					if errors.Is(err, ErrInvalidTheme) || errors.Is(err, ErrInvalidTemperatureUnit) {
						d.Logger().WithError(err).Errorf("Validation failed.")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to update device preferences.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				res, err := ops.Map(Transform)(ops.FixedProvider(model))()
				if err != nil {
					d.Logger().WithError(err).Errorf("Creating REST model.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
			}
		})
	}
}
