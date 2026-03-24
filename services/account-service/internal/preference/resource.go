package preference

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
		rh := server.RegisterHandler(l)(si)
		rih := server.RegisterInputHandler[UpdateRequest](l)(si)

		api.HandleFunc("/preferences", rh("GetPreference", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/preferences/{id}", rih("UpdatePreference", updateHandler(db))).Methods(http.MethodPatch)
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.FindOrCreate(t.Id(), t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to get preference")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())([]RestModel{rest})
		}
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if input.Theme == nil && input.ActiveHouseholdId == nil {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", "at least one field (theme or active_household_id) is required")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				var m Model
				var err error

				if input.Theme != nil {
					m, err = proc.UpdateTheme(id, *input.Theme)
					if err != nil {
						d.Logger().WithError(err).Error("Failed to update theme")
						server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
						return
					}
				}

				if input.ActiveHouseholdId != nil {
					m, err = proc.SetActiveHousehold(id, *input.ActiveHouseholdId)
					if err != nil {
						d.Logger().WithError(err).Error("Failed to set active household")
						server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
						return
					}
				}

				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}
