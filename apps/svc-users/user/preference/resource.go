package preference

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/packages/shared-go/auth"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers all preference-related routes
func InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Preference CRUD endpoints
			router.HandleFunc("/users/me/preferences", server.RegisterHandler(l)(si)("list-preferences", listPreferencesHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/users/me/preferences/{key}", server.RegisterHandler(l)(si)("get-preference", getPreferenceHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/users/me/preferences/{key}", server.RegisterInputHandler[UpdateRequest](l)(si)("update-preference", updatePreferenceHandler(db))).Methods(http.MethodPatch)
			router.HandleFunc("/users/me/preferences/{key}", server.RegisterHandler(l)(si)("delete-preference", deletePreferenceHandler(db))).Methods(http.MethodDelete)
		}
	}
}

// listPreferencesHandler handles GET /users/me/preferences - returns all preferences for the authenticated user
func listPreferencesHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract auth context
			authCtx, ok := auth.FromContext(r.Context())
			if !ok {
				d.Logger().Error("Auth context not found in list preferences handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Get all preferences for user
			processor := NewProcessor(NewGormProvider(db))
			preferences, err := processor.GetAllUserPreferences(authCtx.UserId)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to fetch user preferences")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Transform to REST models
			restModels, err := TransformSlice(preferences)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform preferences")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Marshal using JSON:API format
			server.MarshalResponse[[]RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(restModels)
		}
	}
}

// getPreferenceHandler handles GET /users/me/preferences/{key} - returns a specific preference
func getPreferenceHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract auth context
			authCtx, ok := auth.FromContext(r.Context())
			if !ok {
				d.Logger().Error("Auth context not found in get preference handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Extract key from path
			vars := mux.Vars(r)
			key := vars["key"]
			if key == "" {
				d.Logger().Error("Key not provided in path")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Get preference
			processor := NewProcessor(NewGormProvider(db))
			preference, err := processor.GetUserPreference(authCtx.UserId, key)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					d.Logger().WithError(err).Debugf("Preference not found: %s", key)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				d.Logger().WithError(err).Error("Failed to fetch preference")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Transform to REST model
			restModel, err := Transform(preference)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform preference")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Marshal using JSON:API format
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(restModel)
		}
	}
}

// updatePreferenceHandler handles PATCH /users/me/preferences/{key} - creates or updates a preference
func updatePreferenceHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, model UpdateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract auth context
			authCtx, ok := auth.FromContext(r.Context())
			if !ok {
				d.Logger().Error("Auth context not found in update preference handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Extract key from path
			vars := mux.Vars(r)
			key := vars["key"]
			if key == "" {
				d.Logger().Error("Key not provided in path")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Validate request
			if model.Value == "" {
				d.Logger().Error("Value not provided in request")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Set preference
			processor := NewProcessor(NewGormProvider(db))
			preference, err := processor.SetUserPreference(authCtx.UserId, key, model.Value)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to set preference")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Transform to REST model
			restModel, err := Transform(preference)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform preference")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Marshal using JSON:API format
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(restModel)
		}
	}
}

// deletePreferenceHandler handles DELETE /users/me/preferences/{key} - deletes a preference
func deletePreferenceHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract auth context
			authCtx, ok := auth.FromContext(r.Context())
			if !ok {
				d.Logger().Error("Auth context not found in delete preference handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Extract key from path
			vars := mux.Vars(r)
			key := vars["key"]
			if key == "" {
				d.Logger().Error("Key not provided in path")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Delete preference
			processor := NewProcessor(NewGormProvider(db))
			err := processor.DeleteUserPreference(authCtx.UserId, key)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					d.Logger().WithError(err).Debugf("Preference not found: %s", key)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				d.Logger().WithError(err).Error("Failed to delete preference")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Return 204 No Content on successful deletion
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
