package retention

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	sharedretention "github.com/jtumidanski/home-hub/shared/go/retention"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes wires the public /api/v1 retention endpoints. The fanout
// dependency must be supplied by main.go and is responsible for forwarding
// purge / runs requests to the relevant reaper-owning service.
func InitializeRoutes(db *gorm.DB, fan Fanout) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		patchH := server.RegisterInputHandler[PatchRequest](l)(si)
		purgeH := server.RegisterInputHandler[PurgeRequest](l)(si)

		api.HandleFunc("/retention-policies", rh("GetRetentionPolicies", getPolicies(db))).Methods(http.MethodGet)
		api.HandleFunc("/retention-policies/household/{household_id}", patchH("PatchHouseholdPolicy", patchHousehold(db))).Methods(http.MethodPatch)
		api.HandleFunc("/retention-policies/user", patchH("PatchUserPolicy", patchUser(db))).Methods(http.MethodPatch)
		api.HandleFunc("/retention-policies/purge", purgeH("PurgeRetentionPolicy", purgeHandler(db, fan))).Methods(http.MethodPost)
		api.HandleFunc("/retention-runs", rh("ListRetentionRuns", listRunsHandler(db, fan))).Methods(http.MethodGet)
	}
}

// InitializeInternalRoutes wires the internal endpoints used by reapers in
// other services to load policy. These are NOT under /api/v1 and use a shared
// internal token for auth.
func InitializeInternalRoutes(db *gorm.DB, internalToken string) func(l logrus.FieldLogger, router *mux.Router) {
	return func(l logrus.FieldLogger, router *mux.Router) {
		router.Handle("/internal/retention-policies/overrides",
			internalAuth(internalToken, l)(http.HandlerFunc(getOverridesInternal(db, l))),
		).Methods(http.MethodGet)
	}
}

// internalAuth returns middleware that requires a Bearer token matching the
// configured shared internal-service token. If the token is empty (e.g. local
// dev), the middleware allows the request through.
func internalAuth(expected string, l logrus.FieldLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expected == "" {
				next.ServeHTTP(w, r)
				return
			}
			h := r.Header.Get("Authorization")
			if h != "Bearer "+expected {
				server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getPolicies(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			resolved, err := proc.ResolveAll(t.Id(), t.HouseholdId(), t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("retention: resolve failed")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest := PolicyRest{Id: t.Id()}
			if resolved.Household != nil {
				rest.Household = scopeToRest(resolved.Household)
			}
			if resolved.UserScope != nil {
				rest.User = scopeToRest(resolved.UserScope)
			}
			server.MarshalResponse[PolicyRest](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func scopeToRest(s *ResolvedScope) *PolicyScope {
	out := &PolicyScope{
		Id:         s.ScopeId.String(),
		Categories: make(map[string]CategoryView, len(s.Values)),
	}
	for cat, v := range s.Values {
		out.Categories[string(cat)] = CategoryView{Days: v.Days, Source: v.Source}
	}
	return out
}

func patchHousehold(db *gorm.DB) server.InputHandler[PatchRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input PatchRequest) http.HandlerFunc {
		return server.ParseID("household_id", func(householdID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)

				ok, err := proc.IsHouseholdAdmin(t.Id(), householdID, t.UserId())
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "household not found")
						return
					}
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				if !ok {
					server.WriteError(w, http.StatusForbidden, "Forbidden", "household admin role required")
					return
				}

				patch, err := parseCategoryPatch(input.Categories)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", err.Error())
					return
				}
				if err := proc.ApplyHouseholdPatch(t.Id(), householdID, patch); err != nil {
					if errors.Is(err, ErrInvalidDays) || errors.Is(err, ErrUnknownCategory) || errors.Is(err, ErrScopeMismatch) {
						server.WriteError(w, http.StatusBadRequest, "Bad Request", err.Error())
						return
					}
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				resolved, _ := proc.ResolveAll(t.Id(), householdID, t.UserId())
				rest := PolicyRest{Id: t.Id()}
				if resolved.Household != nil {
					rest.Household = scopeToRest(resolved.Household)
				}
				if resolved.UserScope != nil {
					rest.User = scopeToRest(resolved.UserScope)
				}
				server.MarshalResponse[PolicyRest](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func patchUser(db *gorm.DB) server.InputHandler[PatchRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input PatchRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)

			patch, err := parseCategoryPatch(input.Categories)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Bad Request", err.Error())
				return
			}
			if err := proc.ApplyUserPatch(t.Id(), t.UserId(), patch); err != nil {
				if errors.Is(err, ErrInvalidDays) || errors.Is(err, ErrUnknownCategory) || errors.Is(err, ErrScopeMismatch) {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", err.Error())
					return
				}
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			resolved, _ := proc.ResolveAll(t.Id(), t.HouseholdId(), t.UserId())
			rest := PolicyRest{Id: t.Id()}
			if resolved.Household != nil {
				rest.Household = scopeToRest(resolved.Household)
			}
			if resolved.UserScope != nil {
				rest.User = scopeToRest(resolved.UserScope)
			}
			server.MarshalResponse[PolicyRest](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func parseCategoryPatch(in map[string]*int) (map[sharedretention.Category]*int, error) {
	out := make(map[sharedretention.Category]*int, len(in))
	for k, v := range in {
		c := sharedretention.Category(k)
		if !c.IsKnown() {
			return nil, ErrUnknownCategory
		}
		out[c] = v
	}
	return out, nil
}

func purgeHandler(db *gorm.DB, fan Fanout) server.InputHandler[PurgeRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input PurgeRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			cat := sharedretention.Category(input.Category)
			if !cat.IsKnown() {
				server.WriteError(w, http.StatusBadRequest, "Bad Request", "unknown category")
				return
			}
			scopeKind := sharedretention.ScopeKind(input.Scope)
			if scopeKind != sharedretention.ScopeHousehold && scopeKind != sharedretention.ScopeUser {
				server.WriteError(w, http.StatusBadRequest, "Bad Request", "scope must be household or user")
				return
			}
			if cat.Scope() != scopeKind {
				server.WriteError(w, http.StatusBadRequest, "Bad Request", "category does not match scope")
				return
			}

			var scopeID uuid.UUID
			if scopeKind == sharedretention.ScopeHousehold {
				scopeID = t.HouseholdId()
				ok, err := proc.IsHouseholdAdmin(t.Id(), scopeID, t.UserId())
				if err != nil || !ok {
					server.WriteError(w, http.StatusForbidden, "Forbidden", "household admin role required")
					return
				}
			} else {
				scopeID = t.UserId()
			}

			res, err := fan.Purge(r.Context(), t.Id(), scopeKind, scopeID, cat, input.DryRun)
			if err != nil {
				if errors.Is(err, ErrServiceUnreachable) {
					server.WriteError(w, http.StatusServiceUnavailable, "Service Unavailable", "owning service unreachable")
					return
				}
				if errors.Is(err, ErrRateLimited) {
					server.WriteError(w, http.StatusTooManyRequests, "Too Many Requests", "manual purge rate limit hit")
					return
				}
				d.Logger().WithError(err).Error("retention: purge fanout failed")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			resp := PurgeResponse{
				Id:       res.RunId,
				Category: input.Category,
				Scope:    input.Scope,
				ScopeId:  scopeID.String(),
				Status:   "accepted",
				Scanned:  res.Scanned,
				Deleted:  res.Deleted,
				DryRun:   res.DryRun,
			}
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusAccepted)
			server.MarshalResponse[PurgeResponse](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(resp)
		}
	}
}

func listRunsHandler(db *gorm.DB, fan Fanout) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			q := r.URL.Query()
			runs, err := fan.ListRuns(r.Context(), t.Id(), q.Get("category"), q.Get("trigger"), q.Get("limit"))
			if err != nil {
				d.Logger().WithError(err).Error("retention: list runs failed")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalSliceResponse[RunRest](d.Logger())(w)(c.ServerInformation())(runs)
		}
	}
}

// internal endpoint: returns the override map for a (tenant, scope) pair.
func getOverridesInternal(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		tenantID, err := uuid.Parse(q.Get("tenant_id"))
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "Bad Request", "tenant_id required")
			return
		}
		scopeKind := sharedretention.ScopeKind(q.Get("scope_kind"))
		if scopeKind != sharedretention.ScopeHousehold && scopeKind != sharedretention.ScopeUser {
			server.WriteError(w, http.StatusBadRequest, "Bad Request", "scope_kind required")
			return
		}
		scopeID, err := uuid.Parse(q.Get("scope_id"))
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "Bad Request", "scope_id required")
			return
		}

		proc := NewProcessor(l, r.Context(), db)
		overrides, err := proc.LoadOverrides(tenantID, scopeKind, scopeID)
		if err != nil {
			l.WithError(err).Error("retention: load overrides failed")
			server.WriteError(w, http.StatusInternalServerError, "Error", "")
			return
		}
		out := struct {
			Overrides map[string]int `json:"overrides"`
		}{Overrides: make(map[string]int, len(overrides))}
		for k, v := range overrides {
			out.Overrides[string(k)] = v
		}
		writeJSON(w, http.StatusOK, out)
	}
}
