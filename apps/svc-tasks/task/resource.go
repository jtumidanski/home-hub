package task

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/apps/svc-tasks/user"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
)

// InitializeRoutes registers all task-related routes
func InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Task CRUD endpoints
			router.HandleFunc("/tasks", server.RegisterHandler(l)(si)("get-tasks", listTasksHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/tasks", server.RegisterInputHandler[CreateRequest](l)(si)("create-task", createTaskHandler(db))).Methods(http.MethodPost)
			router.HandleFunc("/tasks/{id}", server.RegisterHandler(l)(si)("get-task", getTaskHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/tasks/{id}", server.RegisterInputHandler[UpdateRequest](l)(si)("update-task", updateTaskHandler(db))).Methods(http.MethodPatch)
			router.HandleFunc("/tasks/{id}", server.RegisterHandler(l)(si)("delete-task", deleteTaskHandler(db))).Methods(http.MethodDelete)
			router.HandleFunc("/tasks/{id}/complete", server.RegisterHandler(l)(si)("complete-task", completeTaskHandler(db))).Methods(http.MethodPost)
		}
	}
}

// IdHandler a handler interface which requires a taskId
type IdHandler func(taskId uuid.UUID) http.HandlerFunc

// ParseId parses the taskId consistently from the request, and provide it to the next handler
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

// listTasksHandler handles GET /tasks
// Supports query params: ?day=YYYY-MM-DD, ?status=incomplete|complete
func listTasksHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context (set by user resolver middleware)
			userId, ok := user.UserIDFromContext(r.Context())
			if !ok {
				d.Logger().Error("User ID not found in /tasks handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Parse query parameters
			query := r.URL.Query()
			dayStr := query.Get("day")
			statusStr := query.Get("status")

			var models []Model
			var err error

			// Filter based on query params
			if dayStr != "" {
				// List by day
				day, parseErr := time.Parse("2006-01-02", dayStr)
				if parseErr != nil {
					d.Logger().WithError(parseErr).Error("Invalid day format")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				models, err = NewProcessor(d.Logger(), r.Context(), db).ListByDay(userId, day)()
			} else if statusStr != "" {
				// List by status
				status := Status(statusStr)
				if !status.IsValid() {
					d.Logger().Error("Invalid status parameter")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				models, err = NewProcessor(d.Logger(), r.Context(), db).ListByStatus(userId, status)()
			} else {
				// List all tasks for user
				models, err = NewProcessor(d.Logger(), r.Context(), db).List(userId)()
			}

			if err != nil {
				d.Logger().WithError(err).Error("error listing tasks")
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

// getTaskHandler handles GET /tasks/:id
func getTaskHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(taskId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in /tasks/:id handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).GetById(taskId, userId)()
				if err != nil {
					if errors.Is(err, ErrTaskNotFound) {
						d.Logger().WithError(err).Error("Task not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Error("Unauthorized access to task")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					d.Logger().WithError(err).Error("Failed to fetch task")
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

// createTaskHandler handles POST /tasks
func createTaskHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context
			userId, ok := user.UserIDFromContext(r.Context())
			if !ok {
				d.Logger().Error("User ID not found in create task handler")
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

			// Parse day from request
			day, err := time.Parse("2006-01-02", req.Day)
			if err != nil {
				d.Logger().WithError(err).Error("Invalid day format")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			input := CreateInput{
				UserId:      userId,
				HouseholdId: householdId,
				Day:         day,
				Title:       req.Title,
				Description: req.Description,
			}

			model, err := NewProcessor(d.Logger(), r.Context(), db).Create(input)()
			if err != nil {
				if errors.Is(err, ErrTitleRequired) || errors.Is(err, ErrTitleEmpty) ||
					errors.Is(err, ErrTitleTooLong) || errors.Is(err, ErrDayRequired) {
					d.Logger().WithError(err).Errorf("Validation failed.")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				d.Logger().WithError(err).Errorf("Failed to create task.")
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

// updateTaskHandler handles PATCH /tasks/:id
func updateTaskHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req UpdateRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(taskId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in update task handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				input := UpdateInput{
					Title:       req.Title,
					Description: req.Description,
				}

				// Parse day if provided
				if req.Day != nil {
					day, err := time.Parse("2006-01-02", *req.Day)
					if err != nil {
						d.Logger().WithError(err).Error("Invalid day format")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					input.Day = &day
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

				model, err := NewProcessor(d.Logger(), r.Context(), db).Update(taskId, userId, input)()
				if err != nil {
					if errors.Is(err, ErrTaskNotFound) {
						d.Logger().WithError(err).Errorf("Task not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Errorf("Unauthorized access to task.")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					if errors.Is(err, ErrTitleRequired) || errors.Is(err, ErrTitleEmpty) ||
						errors.Is(err, ErrTitleTooLong) || errors.Is(err, ErrStatusInvalid) {
						d.Logger().WithError(err).Errorf("Validation failed.")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to update task.")
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

// deleteTaskHandler handles DELETE /tasks/:id
func deleteTaskHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(taskId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in delete task handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				err := NewProcessor(d.Logger(), r.Context(), db).Delete(taskId, userId)
				if err != nil {
					if errors.Is(err, ErrTaskNotFound) {
						d.Logger().WithError(err).Errorf("Task not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Errorf("Unauthorized access to task.")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to delete task.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

// completeTaskHandler handles POST /tasks/:id/complete
func completeTaskHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(taskId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Extract user ID from context
				userId, ok := user.UserIDFromContext(r.Context())
				if !ok {
					d.Logger().Error("User ID not found in complete task handler")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).Complete(taskId, userId)()
				if err != nil {
					if errors.Is(err, ErrTaskNotFound) {
						d.Logger().WithError(err).Errorf("Task not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrUnauthorized) {
						d.Logger().WithError(err).Errorf("Unauthorized access to task.")
						w.WriteHeader(http.StatusForbidden)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to complete task.")
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
