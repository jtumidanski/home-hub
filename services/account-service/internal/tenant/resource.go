package tenant

import (
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
		rih := server.RegisterInputHandler[CreateRequest](l)(si)

		api.HandleFunc("/tenants", rh("ListTenants", listHandler(l, db))).Methods(http.MethodGet)
		api.HandleFunc("/tenants", rih("CreateTenant", createHandler(l, db))).Methods(http.MethodPost)
		api.HandleFunc("/tenants/{id}", rh("GetTenant", getHandler(l, db))).Methods(http.MethodGet)
	}
}

func listHandler(l logrus.FieldLogger, db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			m, err := proc.ByIDProvider(t.Id())()
			if err != nil {
				server.WriteError(w, http.StatusNotFound, "Not Found", "")
				return
			}
			rest, _ := Transform(m)
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())([]RestModel{rest})
		}
	}
}

func createHandler(l logrus.FieldLogger, db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := NewProcessor(logrus.StandardLogger(), r.Context(), db)
			m, err := proc.Create(input.Name)
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}
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
