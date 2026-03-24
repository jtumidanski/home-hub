package reminderdismissal

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/manyminds/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

func (r RestModel) GetName() string       { return "reminder-dismissals" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/reminders/dismissals", rh("CreateReminderDismissal", createHandler(db))).Methods(http.MethodPost)
	}
}

func createHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			body, _ := io.ReadAll(r.Body)
			reminderID, err := extractReminderRelationship(body)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Missing reminder relationship")
				return
			}

			remProc := reminder.NewProcessor(logrus.StandardLogger(), r.Context(), db)
			if err := remProc.Dismiss(reminderID); err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Dismiss Failed", "")
				return
			}

			now := time.Now().UTC()
			e := Entity{
				Id:              uuid.New(),
				TenantId:        t.Id(),
				HouseholdId:     t.HouseholdId(),
				ReminderId:      reminderID,
				CreatedByUserId: t.UserId(),
				CreatedAt:       now,
			}
			db.WithContext(r.Context()).Create(&e)

			rest := RestModel{Id: e.Id, CreatedAt: now}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func extractReminderRelationship(body []byte) (uuid.UUID, error) {
	var env struct {
		Data struct {
			Relationships struct {
				Reminder struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				} `json:"reminder"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return uuid.Nil, err
	}
	if env.Data.Relationships.Reminder.Data.ID == "" {
		return uuid.Nil, errors.New("missing reminder relationship")
	}
	return uuid.Parse(env.Data.Relationships.Reminder.Data.ID)
}
