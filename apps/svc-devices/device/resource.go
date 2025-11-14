package device

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/apps/svc-devices/user"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers all device-related routes
func InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Device CRUD endpoints
			router.HandleFunc("/devices", server.RegisterHandler(l)(si)("get-devices", listDevicesHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/devices", server.RegisterInputHandler[CreateRequest](l)(si)("create-device", createDeviceHandler(db))).Methods(http.MethodPost)
			router.HandleFunc("/devices/{id}", server.RegisterHandler(l)(si)("get-device", getDeviceHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/devices/{id}", server.RegisterInputHandler[UpdateRequest](l)(si)("update-device", updateDeviceHandler(db))).Methods(http.MethodPatch)
			router.HandleFunc("/devices/{id}", server.RegisterHandler(l)(si)("delete-device", deleteDeviceHandler(db))).Methods(http.MethodDelete)
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

// listDevicesHandler handles GET /devices
func listDevicesHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// TODO: Filter by household context from auth
			models, err := GetAll(db)()
			if err != nil {
				d.Logger().WithError(err).Error("error listing devices")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res, err := ops.SliceMap(Transform)(ops.FixedProvider(models))(ops.ParallelMap())()
			if err != nil {
				d.Logger().WithError(err).Errorf("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			server.MarshalResponse[[]RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
		}
	}
}

// getDeviceHandler handles GET /devices/:id
func getDeviceHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(deviceId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				model, err := NewProcessor(d.Logger(), r.Context(), db).GetById(deviceId)()
				if err != nil {
					if errors.Is(err, ErrDeviceNotFound) {
						d.Logger().WithError(err).Error("Device not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to fetch device")
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

// createDeviceHandler handles POST /devices
func createDeviceHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract household ID from user context (set by UserResolverMiddleware)
			userInfo, ok := user.UserInfoFromContext(r.Context())
			if !ok {
				d.Logger().Error("User info not found in context")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if userInfo.HouseholdID == uuid.Nil {
				d.Logger().Error("User has no household associated")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			input := CreateInput{
				Name:        req.Name,
				Type:        req.Type,
				HouseholdId: userInfo.HouseholdID,
			}

			model, err := NewProcessor(d.Logger(), r.Context(), db).Create(input)()
			if err != nil {
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameEmpty) ||
					errors.Is(err, ErrTypeRequired) || errors.Is(err, ErrTypeInvalid) ||
					errors.Is(err, ErrHouseholdRequired) {
					d.Logger().WithError(err).Errorf("Validation failed.")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				d.Logger().WithError(err).Errorf("Failed to create device.")
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
	}
}

// updateDeviceHandler handles PATCH /devices/:id
func updateDeviceHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req UpdateRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(deviceId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				input := UpdateInput{
					Name: req.Name,
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).Update(deviceId, input)()
				if err != nil {
					if errors.Is(err, ErrDeviceNotFound) {
						d.Logger().WithError(err).Errorf("Device not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameEmpty) {
						d.Logger().WithError(err).Errorf("Validation failed.")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to update device.")
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

// deleteDeviceHandler handles DELETE /devices/:id
func deleteDeviceHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(deviceId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				_, err := NewProcessor(d.Logger(), r.Context(), db).Delete(deviceId)()
				if err != nil {
					if errors.Is(err, ErrDeviceNotFound) {
						d.Logger().WithError(err).Error("Device not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to delete device")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
