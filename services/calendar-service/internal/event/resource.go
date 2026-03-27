package event

import (
	"errors"
	"net/http"
	"strings"
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

type SyncConnectionFunc func(conn connection.Model)

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
			conn, err := connProc.ByIDProvider(connID)()
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "Connection not found")
				return
			}
			if conn.UserID() != t.UserId() {
				server.WriteError(w, http.StatusForbidden, "Forbidden", "Not the connection owner")
				return
			}
			if !conn.WriteAccess() {
				server.WriteError(w, http.StatusForbidden, "Write Access Required", "Connection does not have write scope. Please re-authorize.")
				return
			}

			srcProc := source.NewProcessor(d.Logger(), r.Context(), db)
			src, err := srcProc.ByIDProvider(calID)()
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "Calendar not found")
				return
			}
			if src.ConnectionID() != connID {
				server.WriteError(w, http.StatusNotFound, "Not Found", "Calendar not found for this connection")
				return
			}

			accessToken, err := connProc.GetOrRefreshAccessToken(conn, gcClient, enc)
			if err != nil {
				d.Logger().WithError(err).Error("failed to get access token for event creation")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Failed to authenticate with Google")
				return
			}

			evtProc := NewProcessor(d.Logger(), r.Context(), db)
			err = evtProc.CreateOnGoogle(gcClient, accessToken, src.ExternalID(), input)
			if err != nil {
				d.Logger().WithError(err).Error("Google Calendar insert event failed")
				if strings.Contains(err.Error(), "403") {
					server.WriteError(w, http.StatusForbidden, "Google Write Denied", "This calendar does not allow new events")
					return
				}
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Failed to create event on Google Calendar")
				return
			}

			d.Logger().WithField("connection_id", connID.String()).Info("event created on Google Calendar, triggering sync")
			if syncConn != nil {
				syncConn(conn)
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
			conn, err := connProc.ByIDProvider(connID)()
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "Connection not found")
				return
			}
			if conn.UserID() != t.UserId() {
				server.WriteError(w, http.StatusForbidden, "Forbidden", "Not the connection owner")
				return
			}
			if !conn.WriteAccess() {
				server.WriteError(w, http.StatusForbidden, "Write Access Required", "Connection does not have write scope")
				return
			}

			evtProc := NewProcessor(d.Logger(), r.Context(), db)
			evt, err := evtProc.ByID(eventID)
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "Event not found")
				return
			}
			if evt.ConnectionID() != connID {
				server.WriteError(w, http.StatusForbidden, "Forbidden", "Event does not belong to this connection")
				return
			}

			accessToken, err := connProc.GetOrRefreshAccessToken(conn, gcClient, enc)
			if err != nil {
				d.Logger().WithError(err).Error("failed to get access token for event update")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Failed to authenticate with Google")
				return
			}

			err = evtProc.UpdateOnGoogle(gcClient, accessToken, evt, input)
			if err != nil {
				d.Logger().WithError(err).Error("Google Calendar update event failed")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Failed to update event on Google Calendar")
				return
			}

			d.Logger().WithField("event_id", eventID.String()).Info("event updated on Google Calendar, triggering sync")
			if syncConn != nil {
				syncConn(conn)
			}

			updatedEvt, err := evtProc.ByID(eventID)
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Event updated but failed to retrieve")
				return
			}

			rest, err := TransformWithPrivacy(updatedEvt, t.UserId())
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

			connProc := connection.NewProcessor(d.Logger(), r.Context(), db)
			conn, err := connProc.ByIDProvider(connID)()
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "Connection not found")
				return
			}
			if conn.UserID() != t.UserId() {
				server.WriteError(w, http.StatusForbidden, "Forbidden", "Not the connection owner")
				return
			}
			if !conn.WriteAccess() {
				server.WriteError(w, http.StatusForbidden, "Write Access Required", "Connection does not have write scope")
				return
			}

			evtProc := NewProcessor(d.Logger(), r.Context(), db)
			evt, err := evtProc.ByID(eventID)
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "Event not found")
				return
			}
			if evt.ConnectionID() != connID {
				server.WriteError(w, http.StatusForbidden, "Forbidden", "Event does not belong to this connection")
				return
			}

			accessToken, err := connProc.GetOrRefreshAccessToken(conn, gcClient, enc)
			if err != nil {
				d.Logger().WithError(err).Error("failed to get access token for event deletion")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Failed to authenticate with Google")
				return
			}

			scope := r.URL.Query().Get("scope")
			err = evtProc.DeleteOnGoogle(gcClient, accessToken, evt, scope)
			if err != nil {
				d.Logger().WithError(err).Error("Google Calendar delete event failed")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "Failed to delete event on Google Calendar")
				return
			}

			d.Logger().WithField("event_id", eventID.String()).Info("event deleted on Google Calendar, triggering sync")
			if syncConn != nil {
				syncConn(conn)
			}

			w.WriteHeader(http.StatusNoContent)
		}
	}
}
