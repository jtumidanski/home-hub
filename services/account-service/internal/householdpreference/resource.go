package householdpreference

import (
	"encoding/json"
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
		rh := server.RegisterHandler(l)(si)
		rih := server.RegisterInputHandler[UpdateRequest](l)(si)

		api.HandleFunc("/household-preferences", rh("GetHouseholdPreferences", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/household-preferences/{id}", rih("UpdateHouseholdPreferences", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/household-preferences/{id}/kiosk-seeded", rh("MarkKioskSeeded", markKioskSeededHandler(db))).Methods(http.MethodPatch)
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.FindOrCreate(t.Id(), t.UserId(), t.HouseholdId())
			if err != nil {
				d.Logger().WithError(err).Error("failed to get household preferences")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())([]RestModel{rest})
		}
	}
}

// markKioskSeededHandler exposes a sub-route for the write-once-true
// kiosk_dashboard_seeded flag. The body is plain JSON ({"value": true}) rather
// than JSON:API because the existing PATCH /household-preferences/{id} cannot
// distinguish absent-vs-explicit-null on its single mutable attribute (see
// rest.go FIXME). Frontend never sends false; only true is accepted.
func markKioskSeededHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Value bool `json:"value"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !body.Value {
					server.WriteError(w, http.StatusBadRequest, "Invalid Request", "expected {\"value\":true}")
					return
				}
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.MarkKioskSeeded(id)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					d.Logger().WithError(err).Error("failed to mark kiosk seeded")
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("creating REST model")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				// Nil pointer (absent or explicit null) clears the field.
				m, err := proc.SetDefaultDashboard(id, input.DefaultDashboardId)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					d.Logger().WithError(err).Error("failed to set default dashboard")
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("creating REST model")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}
