package remindersnooze

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
	Id              uuid.UUID `json:"-"`
	DurationMinutes int       `json:"durationMinutes"`
	SnoozedUntil    time.Time `json:"snoozedUntil"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (r RestModel) GetName() string       { return "reminder-snoozes" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/reminders/snoozes", rh("CreateReminderSnooze", createHandler(db))).Methods(http.MethodPost)
	}
}

func createHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			body, _ := io.ReadAll(r.Body)
			reminderID, durationMinutes, err := extractSnoozeRequest(body)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", err.Error())
				return
			}

			remProc := reminder.NewProcessor(logrus.StandardLogger(), r.Context(), db)
			snoozedUntil, err := remProc.Snooze(reminderID, durationMinutes)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Snooze Failed", err.Error())
				return
			}

			now := time.Now().UTC()
			e := Entity{
				Id:              uuid.New(),
				TenantId:        t.Id(),
				HouseholdId:     t.HouseholdId(),
				ReminderId:      reminderID,
				DurationMinutes: durationMinutes,
				SnoozedUntil:    snoozedUntil,
				CreatedByUserId: t.UserId(),
				CreatedAt:       now,
			}
			db.WithContext(r.Context()).Create(&e)

			rest := RestModel{Id: e.Id, DurationMinutes: durationMinutes, SnoozedUntil: snoozedUntil, CreatedAt: now}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func extractSnoozeRequest(body []byte) (uuid.UUID, int, error) {
	var env struct {
		Data struct {
			Attributes struct {
				DurationMinutes int `json:"durationMinutes"`
			} `json:"attributes"`
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
		return uuid.Nil, 0, err
	}
	if env.Data.Relationships.Reminder.Data.ID == "" {
		return uuid.Nil, 0, errors.New("missing reminder relationship")
	}
	reminderID, err := uuid.Parse(env.Data.Relationships.Reminder.Data.ID)
	if err != nil {
		return uuid.Nil, 0, err
	}
	return reminderID, env.Data.Attributes.DurationMinutes, nil
}
