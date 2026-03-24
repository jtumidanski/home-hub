package restoration

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rih := server.RegisterInputHandler[CreateRequest](l)(si)
		api.HandleFunc("/tasks/restorations", rih("CreateTaskRestoration", createHandler(db))).Methods(http.MethodPost)
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			taskID, err := uuid.Parse(input.TaskId)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "taskId must be a valid UUID")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			e, err := proc.Create(t.Id(), t.HouseholdId(), taskID, t.UserId())
			if err != nil {
				if errors.Is(err, task.ErrNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Task not found")
				} else if errors.Is(err, task.ErrNotDeleted) {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", "Task is not deleted")
				} else if errors.Is(err, task.ErrRestoreWindow) {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", "Restore window expired")
				} else {
					d.Logger().WithError(err).Error("Failed to restore task")
					server.WriteError(w, http.StatusInternalServerError, "Restore Failed", "")
				}
				return
			}

			rest, err := Transform(e)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}
