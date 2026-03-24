package dismissal

import (
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
		api.HandleFunc("/reminders/dismissals", rih("CreateReminderDismissal", createHandler(db))).Methods(http.MethodPost)
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
			e, err := proc.Create(t.Id(), t.HouseholdId(), reminderID, t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to dismiss reminder")
				server.WriteError(w, http.StatusInternalServerError, "Dismiss Failed", "")
				return
			}

			rest := RestModel{Id: e.Id, CreatedAt: e.CreatedAt}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}
