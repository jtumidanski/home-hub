package dashboard

import (
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
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)
		rihSeed := server.RegisterInputHandler[SeedRequest](l)(si)

		api.HandleFunc("/dashboards", rh("ListDashboards", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/dashboards/order", rh("ReorderDashboards", reorderHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/dashboards/seed", rihSeed("SeedDashboard", seedHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/dashboards", rihCreate("CreateDashboard", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/dashboards/{id}", rh("GetDashboard", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/dashboards/{id}", rihUpdate("UpdateDashboard", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/dashboards/{id}", rh("DeleteDashboard", deleteHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/dashboards/{id}/promote", rh("PromoteDashboard", promoteHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/dashboards/{id}/copy-to-mine", rh("CopyDashboardToMine", copyToMineHandler(db))).Methods(http.MethodPost)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			list, err := proc.List(t.Id(), t.HouseholdId(), t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("list dashboards")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}
			out := make([]RestModel, 0, len(list))
			for _, m := range list {
				rest, err := Transform(m)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				out = append(out, rest)
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(out)
		}
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.GetByID(id, t.Id(), t.HouseholdId(), t.UserId())
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

// --- Stubs: fleshed out in later tasks (Task 27-30). ---

func createHandler(_ *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}
}

func updateHandler(_ *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			}
		})
	}
}

func deleteHandler(_ *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			}
		})
	}
}

func reorderHandler(_ *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}
}

func promoteHandler(_ *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			}
		})
	}
}

func copyToMineHandler(_ *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			}
		})
	}
}

func seedHandler(_ *gorm.DB) server.InputHandler[SeedRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input SeedRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}
}
