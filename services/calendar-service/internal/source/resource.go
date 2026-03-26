package source

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

// ConnectionOwnerChecker validates that a connection belongs to the requesting user.
// Returns the connection's user ID, or an error if not found.
type ConnectionOwnerChecker func(db *gorm.DB, l logrus.FieldLogger, r *http.Request, connID uuid.UUID) (uuid.UUID, error)

func InitializeRoutes(db *gorm.DB, ownerCheck ConnectionOwnerChecker) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rih := server.RegisterInputHandler[ToggleRequest](l)(si)

		api.HandleFunc("/calendar/connections/{id}/calendars", rh("ListSources", listSourcesHandler(db, ownerCheck))).Methods(http.MethodGet)
		api.HandleFunc("/calendar/connections/{id}/calendars/{calId}", rih("ToggleSource", toggleSourceHandler(db, ownerCheck))).Methods(http.MethodPatch)
	}
}

func listSourcesHandler(db *gorm.DB, ownerCheck ConnectionOwnerChecker) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(connID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())

				ownerID, err := ownerCheck(db, d.Logger(), r, connID)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Connection not found")
					return
				}
				if ownerID != t.UserId() {
					server.WriteError(w, http.StatusForbidden, "Forbidden", "Connection belongs to another user")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				models, err := proc.ListByConnection(connID)
				if err != nil {
					d.Logger().WithError(err).Error("failed to list sources")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}

				rest, err := TransformSlice(models)
				if err != nil {
					d.Logger().WithError(err).Error("transforming sources")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
			}
		})
	}
}

func toggleSourceHandler(db *gorm.DB, ownerCheck ConnectionOwnerChecker) server.InputHandler[ToggleRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input ToggleRequest) http.HandlerFunc {
		return server.ParseID("id", func(connID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				calID := mux.Vars(r)["calId"]

				ownerID, err := ownerCheck(db, d.Logger(), r, connID)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Connection not found")
					return
				}
				if ownerID != t.UserId() {
					server.WriteError(w, http.StatusForbidden, "Forbidden", "Connection belongs to another user")
					return
				}

				sourceID, err := uuid.Parse(calID)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid ID", "Invalid calendar ID")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				if err := proc.ToggleVisibility(sourceID, input.Visible); err != nil {
					d.Logger().WithError(err).Error("failed to toggle source visibility")
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}

				updated, err := proc.ByIDProvider(sourceID)()
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Source not found")
					return
				}

				rest, err := Transform(updated)
				if err != nil {
					d.Logger().WithError(err).Error("transforming source")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}
