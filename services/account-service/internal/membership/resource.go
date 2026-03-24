package membership

import (
	"encoding/json"
	"errors"
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

		api.HandleFunc("/memberships", rh("ListMemberships", listHandler(l, db))).Methods(http.MethodGet)
		api.HandleFunc("/memberships", rh("CreateMembership", createHandler(l, db))).Methods(http.MethodPost)
		api.HandleFunc("/memberships/{id}", rh("UpdateMembership", updateHandler(l, db))).Methods(http.MethodPatch)
		api.HandleFunc("/memberships/{id}", rh("DeleteMembership", deleteHandler(l, db))).Methods(http.MethodDelete)
	}
}

func listHandler(l logrus.FieldLogger, db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			models, err := proc.ByUserAndTenantProvider(t.UserId(), t.Id())()
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			var rest []RestModel
			for _, m := range models {
				rm, _ := Transform(m)
				rest = append(rest, rm)
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func createHandler(l logrus.FieldLogger, db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			body, _ := io.ReadAll(r.Body)
			var input CreateRequest
			if err := jsonapi.Unmarshal(body, &input); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "")
				return
			}

			householdID, userID, err := extractRelationships(body)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "Missing relationships")
				return
			}

			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			m, err := proc.Create(t.Id(), householdID, userID, input.Role)
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}
			rest, _ := Transform(m)
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
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
			var input CreateRequest
			if err := jsonapi.Unmarshal(body, &input); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "")
				return
			}

			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			m, err := proc.UpdateRole(id, input.Role)
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
				return
			}
			rest, _ := Transform(m)
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func deleteHandler(l logrus.FieldLogger, db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			if err := proc.Delete(id); err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// extractRelationships extracts household and user IDs from a JSON:API request body.
func extractRelationships(body []byte) (householdID, userID uuid.UUID, err error) {
	var env struct {
		Data struct {
			Relationships map[string]struct {
				Data struct {
					ID string `json:"id"`
				} `json:"data"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err = json.Unmarshal(body, &env); err != nil {
		return
	}
	hhRel, ok := env.Data.Relationships["household"]
	if !ok || hhRel.Data.ID == "" {
		err = errors.New("missing household relationship")
		return
	}
	householdID, err = uuid.Parse(hhRel.Data.ID)
	if err != nil {
		return
	}
	userRel, ok := env.Data.Relationships["user"]
	if !ok || userRel.Data.ID == "" {
		err = errors.New("missing user relationship")
		return
	}
	userID, err = uuid.Parse(userRel.Data.ID)
	return
}
