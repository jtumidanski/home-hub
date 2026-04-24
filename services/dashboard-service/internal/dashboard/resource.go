package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/layout"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// writeLayoutError maps a layout.ValidationError onto a JSON:API 422 with the
// stable layout code + source pointer so the UI can highlight the offending
// widget/field.
func writeLayoutError(w http.ResponseWriter, ve layout.ValidationError) {
	server.WriteJSONAPIError(w, http.StatusUnprocessableEntity, string(ve.Code), "Layout validation failed", ve.Message, ve.Pointer)
}

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

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), t.HouseholdId(), t.UserId(), CreateAttrs{
				Name:      input.Name,
				Scope:     input.Scope,
				Layout:    input.Layout,
				SortOrder: input.SortOrder,
			})
			if err != nil {
				var ve layout.ValidationError
				if errors.As(err, &ve) {
					writeLayoutError(w, ve)
					return
				}
				if errors.Is(err, ErrInvalidScope) {
					server.WriteJSONAPIError(w, http.StatusBadRequest, "dashboard.invalid_scope", "Invalid scope", err.Error(), "/data/attributes/scope")
					return
				}
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) {
					server.WriteJSONAPIError(w, http.StatusUnprocessableEntity, "dashboard.name_invalid", "Invalid name", err.Error(), "/data/attributes/name")
					return
				}
				d.Logger().WithError(err).Error("create dashboard")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}
			rest, err := Transform(m)
			if err != nil {
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Update(id, t.Id(), t.HouseholdId(), t.UserId(), UpdateAttrs{
					Name:      input.Name,
					Layout:    input.Layout,
					SortOrder: input.SortOrder,
				})
				if err != nil {
					var ve layout.ValidationError
					if errors.As(err, &ve) {
						writeLayoutError(w, ve)
						return
					}
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					if errors.Is(err, ErrForbidden) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) {
						server.WriteJSONAPIError(w, http.StatusUnprocessableEntity, "dashboard.name_invalid", "Invalid name", err.Error(), "/data/attributes/name")
						return
					}
					d.Logger().WithError(err).Error("update dashboard")
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

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				if err := proc.Delete(id, t.Id(), t.HouseholdId(), t.UserId()); err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					if errors.Is(err, ErrForbidden) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					d.Logger().WithError(err).Error("delete dashboard")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func reorderHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			// Reorder uses a plain JSON body, not JSON:API-wrapped, since it is a
			// bulk action rather than a single-resource mutation.
			var req ReorderRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "could not parse reorder body")
				return
			}

			pairs := make([]ReorderPair, 0, len(req.Entries))
			for _, e := range req.Entries {
				id, err := uuid.Parse(e.ID)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid ID", "reorder entry has invalid uuid")
					return
				}
				pairs = append(pairs, ReorderPair{ID: id, SortOrder: e.SortOrder})
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			list, err := proc.Reorder(t.Id(), t.HouseholdId(), t.UserId(), pairs)
			if err != nil {
				if errors.Is(err, ErrMixedScope) {
					server.WriteJSONAPIError(w, http.StatusBadRequest, "dashboard.mixed_scope", "Mixed scope", err.Error(), "")
					return
				}
				if errors.Is(err, ErrNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "")
					return
				}
				d.Logger().WithError(err).Error("reorder dashboards")
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

func promoteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Promote(id, t.Id(), t.HouseholdId(), t.UserId())
				if err != nil {
					if errors.Is(err, ErrAlreadyHousehold) {
						server.WriteJSONAPIError(w, http.StatusConflict, "dashboard.already_household", "Already household-scoped", err.Error(), "")
						return
					}
					if errors.Is(err, ErrForbidden) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					d.Logger().WithError(err).Error("promote dashboard")
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

func copyToMineHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.CopyToMine(id, t.Id(), t.HouseholdId(), t.UserId())
				if err != nil {
					if errors.Is(err, ErrNotCopyable) {
						server.WriteJSONAPIError(w, http.StatusBadRequest, "dashboard.not_copyable", "Not copyable", err.Error(), "")
						return
					}
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					d.Logger().WithError(err).Error("copy dashboard to mine")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
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
