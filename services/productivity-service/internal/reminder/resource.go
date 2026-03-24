package reminder

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

		api.HandleFunc("/reminders", rh("ListReminders", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/reminders", rih("CreateReminder", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/reminders/{id}", rh("GetReminder", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/reminders/{id}", rih("UpdateReminder", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/reminders/{id}", rh("DeleteReminder", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.AllProvider()()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list reminders")
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
			scheduledFor, err := time.Parse(time.RFC3339, input.ScheduledFor)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Date", "scheduledFor must be ISO-8601")
				return
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), t.HouseholdId(), input.Title, input.Notes, scheduledFor)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to create reminder")
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
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.ByIDProvider(id)()
			if err != nil {
				d.Logger().WithError(err).Error("Reminder not found")
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
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}
			scheduledFor, err := time.Parse(time.RFC3339, input.ScheduledFor)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Date", "")
				return
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Update(id, input.Title, input.Notes, scheduledFor)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to update reminder")
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
	}
}

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			if err := proc.Delete(id); err != nil {
				d.Logger().WithError(err).Error("Failed to delete reminder")
				server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
