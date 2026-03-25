package snooze

import (
	"errors"
	"net/http"

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
		rih := server.RegisterInputHandler[CreateRequest](l)(si)
		api.HandleFunc("/reminders/snoozes", rih("CreateReminderSnooze", createHandler(db))).Methods(http.MethodPost)
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			reminderID, err := uuid.Parse(input.ReminderId)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "reminderId must be a valid UUID")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), t.HouseholdId(), reminderID, t.UserId(), input.DurationMinutes)
			if err != nil {
				if errors.Is(err, ErrReminderIDRequired) || errors.Is(err, ErrCreatedByRequired) || errors.Is(err, ErrDurationMinutesRequired) {
					server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Invalid snooze parameters")
					return
				}
				d.Logger().WithError(err).Error("Failed to snooze reminder")
				server.WriteError(w, http.StatusInternalServerError, "Snooze Failed", "")
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
