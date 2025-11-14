package reminder

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/apps/svc-reminders/user"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers all reminder-related routes
func InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Reminder CRUD endpoints
			router.HandleFunc("/reminders", server.RegisterHandler(l)(si)("get-reminders", listRemindersHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/reminders/household", server.RegisterHandler(l)(si)("get-household-reminders", listHouseholdRemindersHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/reminders", server.RegisterInputHandler[CreateRequest](l)(si)("create-reminder", createReminderHandler(db))).Methods(http.MethodPost)
			router.HandleFunc("/reminders/{id}", server.RegisterHandler(l)(si)("get-reminder", getReminderHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/reminders/{id}", server.RegisterInputHandler[UpdateRequest](l)(si)("update-reminder", updateReminderHandler(db))).Methods(http.MethodPatch)
			router.HandleFunc("/reminders/{id}", server.RegisterHandler(l)(si)("delete-reminder", deleteReminderHandler(db))).Methods(http.MethodDelete)
			router.HandleFunc("/reminders/{id}/snooze", server.RegisterInputHandler[SnoozeRequest](l)(si)("snooze-reminder", snoozeReminderHandler(db))).Methods(http.MethodPost)
			router.HandleFunc("/reminders/{id}/dismiss", server.RegisterHandler(l)(si)("dismiss-reminder", dismissReminderHandler(db))).Methods(http.MethodPost)
		}
	}
}

// IdHandler a handler interface which requires a reminderId
type IdHandler func(reminderId uuid.UUID) http.HandlerFunc

// ParseId parses the reminderId consistently from the request, and provide it to the next handler
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

// listRemindersHandler handles GET /reminders
// Supports query params: ?status=active|snoozed|dismissed
func listRemindersHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context (set by user resolver middleware)
			userId, ok := user.UserIDFromContext(r.Context())
			if !ok {
				d.Logger().Error("User ID not found in /reminders handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Parse query parameters
			query := r.URL.Query()
			statusStr := query.Get("status")

			var models []Model
			var err error

			// Filter based on query params
			if statusStr != "" {
				// List by status
				status := Status(statusStr)
				if !status.IsValid() {
					d.Logger().Error("Invalid status parameter")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				models, err = NewProcessor(d.Logger(), r.Context(), db).ListByStatus(userId, status)()
			} else {
				// List all reminders for user
				models, err = NewProcessor(d.Logger(), r.Context(), db).List(userId)()
			}

			if err != nil {
				d.Logger().WithError(err).Error("error listing reminders")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Errorf("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			server.MarshalResponse[[]RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
		}
	}
}

// listHouseholdRemindersHandler handles GET /reminders/household
func listHouseholdRemindersHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract household ID from context
			householdId, ok := user.HouseholdIDFromContext(r.Context())
			if !ok {
				d.Logger().Error("Household ID not found in /reminders/household handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			models, err := NewProcessor(d.Logger(), r.Context(), db).ListByHousehold(householdId)()
			if err != nil {
				d.Logger().WithError(err).Error("error listing household reminders")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Errorf("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			server.MarshalResponse[[]RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
		}
	}
}

// getReminderHandler handles GET /reminders/:id
func getReminderHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(reminderId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in /reminders/:id handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).GetById(reminderId, userId)()
				if err != nil {
					if errors.Is(err, ErrReminderNotFound) {
						d.Logger().WithError(err).Error("Reminder not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Error("Unauthorized access to reminder")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					d.Logger().WithError(err).Error("Failed to fetch reminder")
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

// createReminderHandler handles POST /reminders
func createReminderHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context
			userId, ok := user.UserIDFromContext(r.Context())
			if !ok {
				d.Logger().Error("User ID not found in create reminder handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Extract household ID from context (provided by users service)
			householdId, ok := user.HouseholdIDFromContext(r.Context())
			if !ok || householdId == uuid.Nil {
				d.Logger().Error("Household ID not found in context - user not associated with household")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Parse remindAt from request
			remindAt, err := time.Parse(time.RFC3339, req.RemindAt)
			if err != nil {
				d.Logger().WithError(err).Error("Invalid remindAt format")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			input := CreateInput{
				UserId:      userId,
				HouseholdId: householdId,
				Name:        req.Name,
				Description: req.Description,
				RemindAt:    remindAt,
			}

			model, err := NewProcessor(d.Logger(), r.Context(), db).Create(input)()
			if err != nil {
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameEmpty) ||
					errors.Is(err, ErrNameTooLong) || errors.Is(err, ErrRemindAtRequired) {
					d.Logger().WithError(err).Errorf("Validation failed.")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				d.Logger().WithError(err).Errorf("Failed to create reminder.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res, err := ops.Map(Transform)(ops.FixedProvider(model))()
			if err != nil {
				d.Logger().WithError(err).Errorf("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
		}
	}
}

// updateReminderHandler handles PATCH /reminders/:id
func updateReminderHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req UpdateRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(reminderId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in update reminder handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				input := UpdateInput{
					Name:        req.Name,
					Description: req.Description,
				}

				// Parse remindAt if provided
				if req.RemindAt != nil {
					remindAt, err := time.Parse(time.RFC3339, *req.RemindAt)
					if err != nil {
						d.Logger().WithError(err).Error("Invalid remindAt format")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					input.RemindAt = &remindAt
				}

				// Parse status if provided
				if req.Status != nil {
					status := Status(*req.Status)
					if !status.IsValid() {
						d.Logger().Error("Invalid status value")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					input.Status = &status
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).Update(reminderId, userId, input)()
				if err != nil {
					if errors.Is(err, ErrReminderNotFound) {
						d.Logger().WithError(err).Errorf("Reminder not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Errorf("Unauthorized access to reminder.")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameEmpty) ||
						errors.Is(err, ErrNameTooLong) || errors.Is(err, ErrStatusInvalid) {
						d.Logger().WithError(err).Errorf("Validation failed.")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to update reminder.")
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

// deleteReminderHandler handles DELETE /reminders/:id
func deleteReminderHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(reminderId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in delete reminder handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				err := NewProcessor(d.Logger(), r.Context(), db).Delete(reminderId, userId)
				if err != nil {
					if errors.Is(err, ErrReminderNotFound) {
						d.Logger().WithError(err).Errorf("Reminder not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Errorf("Unauthorized access to reminder.")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to delete reminder.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

// snoozeReminderHandler handles POST /reminders/:id/snooze
func snoozeReminderHandler(db *gorm.DB) server.InputHandler[SnoozeRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req SnoozeRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(reminderId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in snooze reminder handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Parse new remindAt from request
				newRemindAt, err := time.Parse(time.RFC3339, req.RemindAt)
				if err != nil {
					d.Logger().WithError(err).Error("Invalid remindAt format")
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).Snooze(reminderId, userId, newRemindAt)()
				if err != nil {
					if errors.Is(err, ErrReminderNotFound) {
						d.Logger().WithError(err).Errorf("Reminder not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Errorf("Unauthorized access to reminder.")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to snooze reminder.")
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

// dismissReminderHandler handles POST /reminders/:id/dismiss
func dismissReminderHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(reminderId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in dismiss reminder handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).Dismiss(reminderId, userId)()
				if err != nil {
					if errors.Is(err, ErrReminderNotFound) {
						d.Logger().WithError(err).Errorf("Reminder not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Errorf("Unauthorized access to reminder.")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to dismiss reminder.")
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
