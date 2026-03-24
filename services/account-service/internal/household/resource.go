package household

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
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)

		api.HandleFunc("/households", rh("ListHouseholds", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/households", rihCreate("CreateHousehold", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/households/{id}", rh("GetHousehold", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/households/{id}", rihUpdate("UpdateHousehold", updateHandler(db))).Methods(http.MethodPatch)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.AllProvider()()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list households")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST models")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.CreateWithOwner(t.Id(), t.UserId(), input.Name, input.Timezone, input.Units)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to create household")
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}
			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.ByIDProvider(id)()
			if err != nil {
				d.Logger().WithError(err).Error("Household not found")
				server.WriteError(w, http.StatusNotFound, "Not Found", "")
				return
			}
			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := uuid.Parse(mux.Vars(r)["id"])
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid ID", "")
				return
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Update(id, input.Name, input.Timezone, input.Units)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to update household")
				server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
				return
			}
			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}
