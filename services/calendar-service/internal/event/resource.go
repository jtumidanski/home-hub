package event

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/connection"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/crypto"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/source"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/calendar/events", rh("ListEvents", listEventsHandler(db))).Methods(http.MethodGet)
	}
}

func InitializeMutationRoutes(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, syncConn SyncConnectionFunc) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rih := server.RegisterInputHandler[CreateEventRequest](l)(si)
		ruih := server.RegisterInputHandler[UpdateEventRequest](l)(si)
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/calendar/connections/{connId}/calendars/{calId}/events", rih("CreateEvent", createEventHandler(db, gcClient, enc, syncConn))).Methods(http.MethodPost)
		api.HandleFunc("/calendar/connections/{connId}/events/{eventId}", ruih("UpdateEvent", updateEventHandler(db, gcClient, enc, syncConn))).Methods(http.MethodPatch)
		api.HandleFunc("/calendar/connections/{connId}/events/{eventId}", rh("DeleteEvent", deleteEventHandler(db, gcClient, enc, syncConn))).Methods(http.MethodDelete)
	}
}

func listEventsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			now := time.Now().UTC()
			startStr := r.URL.Query().Get("start")
			endStr := r.URL.Query().Get("end")

			var start, end time.Time
			var err error

			if startStr != "" {
				start, err = time.Parse(time.RFC3339, startStr)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Parameter", "Invalid start date format, use ISO 8601")
					return
				}
			} else {
				start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			}

			if endStr != "" {
				end, err = time.Parse(time.RFC3339, endStr)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Parameter", "Invalid end date format, use ISO 8601")
					return
				}
			} else {
				end = start.AddDate(0, 0, 7)
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.QueryByHouseholdAndTimeRange(t.HouseholdId(), start, end)
			if err != nil {
				if errors.Is(err, ErrRangeTooLarge) {
					server.WriteError(w, http.StatusBadRequest, "Range Too Large", "Maximum query range is 90 days")
					return
				}
				d.Logger().WithError(err).Error("failed to query events")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}

			rest, err := TransformSliceWithPrivacy(models, t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("transforming events")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func createEventHandler(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, syncConn SyncConnectionFunc) server.InputHandler[CreateEventRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateEventRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			connIDStr := mux.Vars(r)["connId"]
			calIDStr := mux.Vars(r)["calId"]

			connID, err := uuid.Parse(connIDStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "Invalid connection ID")
				return
			}
			calID, err := uuid.Parse(calIDStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "Invalid calendar ID")
				return
			}

			if input.Title == "" {
				server.WriteError(w, http.StatusUnprocessableEntity, "Validation Error", "Title is required")
				return
			}

			connProc := connection.NewProcessor(d.Logger(), r.Context(), db)
			srcProc := source.NewProcessor(d.Logger(), r.Context(), db)
			proc := NewMutationProcessor(d.Logger(), r.Context(), db, connProc, srcProc)
			err = proc.CreateEventOnGoogle(connID, calID, t.UserId(), input, gcClient, enc, syncConn)
			if err != nil {
				handleMutationError(d, w, err, "create")
				return
			}

			w.WriteHeader(http.StatusCreated)
		}
	}
}

func updateEventHandler(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, syncConn SyncConnectionFunc) server.InputHandler[UpdateEventRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateEventRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			connIDStr := mux.Vars(r)["connId"]
			eventIDStr := mux.Vars(r)["eventId"]

			connID, err := uuid.Parse(connIDStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "Invalid connection ID")
				return
			}
			eventID, err := uuid.Parse(eventIDStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "Invalid event ID")
				return
			}

			connProc := connection.NewProcessor(d.Logger(), r.Context(), db)
			srcProc := source.NewProcessor(d.Logger(), r.Context(), db)
			proc := NewMutationProcessor(d.Logger(), r.Context(), db, connProc, srcProc)
			updated, err := proc.UpdateEventOnGoogle(connID, eventID, t.UserId(), input, gcClient, enc, syncConn)
			if err != nil {
				handleMutationError(d, w, err, "update")
				return
			}

			rest, err := TransformWithPrivacy(updated, t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("transforming updated event")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func deleteEventHandler(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, syncConn SyncConnectionFunc) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			connIDStr := mux.Vars(r)["connId"]
			eventIDStr := mux.Vars(r)["eventId"]

			connID, err := uuid.Parse(connIDStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "Invalid connection ID")
				return
			}
			eventID, err := uuid.Parse(eventIDStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "Invalid event ID")
				return
			}

			scope := r.URL.Query().Get("scope")

			connProc := connection.NewProcessor(d.Logger(), r.Context(), db)
			srcProc := source.NewProcessor(d.Logger(), r.Context(), db)
			proc := NewMutationProcessor(d.Logger(), r.Context(), db, connProc, srcProc)
			err = proc.DeleteEventOnGoogle(connID, eventID, t.UserId(), scope, gcClient, enc, syncConn)
			if err != nil {
				handleMutationError(d, w, err, "delete")
				return
			}

			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleMutationError(d *server.HandlerDependency, w http.ResponseWriter, err error, operation string) {
	switch {
	case errors.Is(err, ErrConnectionNotFound):
		server.WriteError(w, http.StatusNotFound, "Not Found", "Connection not found")
	case errors.Is(err, ErrNotOwner):
		server.WriteError(w, http.StatusForbidden, "Forbidden", "Not the connection owner")
	case errors.Is(err, ErrNoWriteAccess):
		server.WriteError(w, http.StatusForbidden, "Write Access Required", "Connection does not have write scope. Please re-authorize.")
	case errors.Is(err, ErrSourceNotFound), errors.Is(err, ErrSourceMismatch):
		server.WriteError(w, http.StatusNotFound, "Not Found", "Calendar not found for this connection")
	case errors.Is(err, ErrEventNotFound), errors.Is(err, ErrEventMismatch):
		server.WriteError(w, http.StatusNotFound, "Not Found", "Event not found")
	case errors.Is(err, ErrGoogleWriteDenied):
		server.WriteError(w, http.StatusForbidden, "Google Write Denied", "This calendar does not allow new events")
	case errors.Is(err, ErrAuthFailed):
		server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Failed to authenticate with Google")
	default:
		d.Logger().WithError(err).Errorf("Google Calendar %s event failed", operation)
		server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Failed to "+operation+" event on Google Calendar")
	}
}
