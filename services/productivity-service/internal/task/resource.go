package task

import (
	"net/http"
	"time"

	"github.com/google/uuid"
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
		rih := server.RegisterInputHandler[CreateRequest](l)(si)
		rihu := server.RegisterInputHandler[UpdateRequest](l)(si)

		api.HandleFunc("/tasks", rh("ListTasks", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/tasks", rih("CreateTask", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/tasks/{id}", rh("GetTask", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/tasks/{id}", rihu("UpdateTask", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/tasks/{id}", rh("DeleteTask", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.AllProvider(false)()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list tasks")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST models")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			var dueOn *time.Time
			if input.DueOn != nil {
				parsed, err := time.Parse("2006-01-02", *input.DueOn)
				if err == nil {
					dueOn = &parsed
				}
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), t.HouseholdId(), input.Title, input.Notes, dueOn, input.RolloverEnabled)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to create task")
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}
			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.ByIDProvider(id)()
				if err != nil {
					d.Logger().WithError(err).Error("Task not found")
					server.WriteError(w, http.StatusNotFound, "Not Found", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				var dueOn *time.Time
				if input.DueOn != nil {
					parsed, err := time.Parse("2006-01-02", *input.DueOn)
					if err == nil {
						dueOn = &parsed
					}
				}
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Update(id, input.Title, input.Notes, input.Status, dueOn, input.RolloverEnabled, t.UserId())
				if err != nil {
					d.Logger().WithError(err).Error("Failed to update task")
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				if err := proc.Delete(id); err != nil {
					d.Logger().WithError(err).Error("Failed to delete task")
					server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
