package taskrestoration

import (
	"encoding/json"
	"errors"
	"io"
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

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	CreatedAt string    `json:"createdAt"`
}

func (r RestModel) GetName() string       { return "task-restorations" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/tasks/restorations", rh("CreateTaskRestoration", createHandler(db))).Methods(http.MethodPost)
	}
}

func createHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			body, _ := io.ReadAll(r.Body)
			taskID, err := extractTaskRelationship(body)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Missing task relationship")
				return
			}

			taskProc := task.NewProcessor(logrus.StandardLogger(), r.Context(), db)
			if err := taskProc.Restore(taskID); err != nil {
				if errors.Is(err, task.ErrNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Task not found")
				} else if errors.Is(err, task.ErrNotDeleted) {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", "Task is not deleted")
				} else if errors.Is(err, task.ErrRestoreWindow) {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", "Restore window expired")
				} else {
					server.WriteError(w, http.StatusInternalServerError, "Restore Failed", "")
				}
				return
			}

			e, err := create(db.WithContext(r.Context()), t.Id(), t.HouseholdId(), taskID, t.UserId())
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Record Failed", "")
				return
			}

			rest := RestModel{Id: e.Id, CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func extractTaskRelationship(body []byte) (uuid.UUID, error) {
	var env struct {
		Data struct {
			Relationships struct {
				Task struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				} `json:"task"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return uuid.Nil, err
	}
	if env.Data.Relationships.Task.Data.ID == "" {
		return uuid.Nil, errors.New("missing task relationship")
	}
	return uuid.Parse(env.Data.Relationships.Task.Data.ID)
}
