package preference

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/manyminds/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/preferences", rh("GetPreference", getHandler(l, db))).Methods(http.MethodGet)
		api.HandleFunc("/preferences/{id}", rh("UpdatePreference", updateHandler(l, db))).Methods(http.MethodPatch)
	}
}

func getHandler(l logrus.FieldLogger, db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			m, err := proc.FindOrCreate(t.Id(), t.UserId())
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest, _ := Transform(m)
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())([]RestModel{rest})
		}
	}
}

func updateHandler(l logrus.FieldLogger, db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}

			body, _ := io.ReadAll(r.Body)

			var input UpdateRequest
			if err := jsonapi.Unmarshal(body, &input); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "")
				return
			}

			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			var m Model

			if input.Theme != nil {
				m, err = proc.UpdateTheme(id, *input.Theme)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}
			}

			hhID, relErr := extractActiveHouseholdRelationship(body)
			if relErr == nil {
				m, err = proc.SetActiveHousehold(id, hhID)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}
			}

			rest, _ := Transform(m)
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

// extractActiveHouseholdRelationship extracts the activeHousehold relationship ID from the request body.
func extractActiveHouseholdRelationship(body []byte) (uuid.UUID, error) {
	var env struct {
		Data struct {
			Relationships map[string]struct {
				Data struct {
					ID string `json:"id"`
				} `json:"data"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return uuid.UUID{}, err
	}
	hhRel, ok := env.Data.Relationships["activeHousehold"]
	if !ok || hhRel.Data.ID == "" {
		return uuid.UUID{}, json.Unmarshal(nil, nil) // return error
	}
	return uuid.Parse(hhRel.Data.ID)
}
