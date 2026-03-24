package household

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/manyminds/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rih := server.RegisterInputHandler[CreateRequest](l)(si)

		api.HandleFunc("/households", rh("ListHouseholds", listHandler(l, db))).Methods(http.MethodGet)
		api.HandleFunc("/households", rih("CreateHousehold", createHandler(l, db))).Methods(http.MethodPost)
		api.HandleFunc("/households/{id}", rh("GetHousehold", getHandler(l, db))).Methods(http.MethodGet)
		api.HandleFunc("/households/{id}", rih("UpdateHousehold", updateHandler(l, db))).Methods(http.MethodPatch)
	}
}

func listHandler(l logrus.FieldLogger, db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			models, err := proc.ByTenantIDProvider(t.Id())()
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

func createHandler(l logrus.FieldLogger, db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			m, err := proc.Create(t.Id(), input.Name, input.Timezone, input.Units)
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}

			// Auto-create owner membership
			memProc := membership.NewProcessor(logrus.StandardLogger(), r.Context(), db)
			memProc.Create(t.Id(), m.Id(), t.UserId(), "owner")

			rest, _ := Transform(m)
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func getHandler(l logrus.FieldLogger, db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			m, err := proc.ByIDProvider(id)()
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "")
				return
			}
			rest, _ := Transform(m)
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func updateHandler(l logrus.FieldLogger, db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			m, err := proc.Update(id, input.Name, input.Timezone, input.Units)
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
				return
			}
			rest, _ := Transform(m)
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}
