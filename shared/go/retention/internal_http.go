package retention

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MountInternalEndpoints registers POST /internal/retention/purge and
// GET /internal/retention/runs on the supplied router. The bearer token, if
// non-empty, is required on every request. The reaper supplies the handlers,
// audit DB, and the per-(tenant, category) rate limiter.
func MountInternalEndpoints(router *mux.Router, reaper *Reaper, token string, l logrus.FieldLogger) {
	rl := newRateLimiter(60 * time.Second)

	router.Handle("/internal/retention/purge",
		bearer(token)(http.HandlerFunc(handlePurge(reaper, rl, l))),
	).Methods(http.MethodPost)

	router.Handle("/internal/retention/runs",
		bearer(token)(http.HandlerFunc(handleListRuns(reaper, l))),
	).Methods(http.MethodGet)
}

func bearer(expected string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expected == "" {
				next.ServeHTTP(w, r)
				return
			}
			if r.Header.Get("Authorization") != "Bearer "+expected {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type purgeReq struct {
	TenantId  string `json:"tenant_id"`
	ScopeKind string `json:"scope_kind"`
	ScopeId   string `json:"scope_id"`
	Category  string `json:"category"`
	DryRun    bool   `json:"dry_run"`
}

type purgeResp struct {
	RunId      string `json:"run_id"`
	Scanned    int    `json:"scanned"`
	Deleted    int    `json:"deleted"`
	DryRun     bool   `json:"dry_run"`
	DurationMs int64  `json:"duration_ms"`
}

func handlePurge(reaper *Reaper, rl *rateLimiter, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req purgeReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		tenantID, err := uuid.Parse(req.TenantId)
		if err != nil {
			http.Error(w, "tenant_id required", http.StatusBadRequest)
			return
		}
		scopeID, err := uuid.Parse(req.ScopeId)
		if err != nil {
			http.Error(w, "scope_id required", http.StatusBadRequest)
			return
		}
		cat := Category(req.Category)
		handler := reaper.HandlerFor(cat)
		if handler == nil {
			http.Error(w, "category not owned by this service", http.StatusBadRequest)
			return
		}

		if !rl.allow(tenantID, cat) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}

		started := time.Now()
		scope := Scope{TenantId: tenantID, Kind: ScopeKind(req.ScopeKind), ScopeId: scopeID}
		rec := reaper.RunOne(r.Context(), handler, scope, TriggerManual, req.DryRun)
		dur := time.Since(started)

		// Translate the recorded error → HTTP status.
		if rec.Error != "" {
			if rec.Error == ErrPolicyUnavailable.Error() {
				http.Error(w, "policy unavailable", http.StatusServiceUnavailable)
				return
			}
			if rec.Error == errLockBusy.Error() {
				http.Error(w, "lock busy", http.StatusConflict)
				return
			}
			http.Error(w, rec.Error, http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, purgeResp{
			RunId:      rec.Id.String(),
			Scanned:    rec.Scanned,
			Deleted:    rec.Deleted,
			DryRun:     rec.DryRun,
			DurationMs: dur.Milliseconds(),
		})
	}
}

type runResp struct {
	Service    string `json:"service"`
	RunId      string `json:"run_id"`
	TenantId   string `json:"tenant_id"`
	ScopeKind  string `json:"scope_kind"`
	ScopeId    string `json:"scope_id"`
	Category   string `json:"category"`
	Trigger    string `json:"trigger"`
	DryRun     bool   `json:"dry_run"`
	Scanned    int    `json:"scanned"`
	Deleted    int    `json:"deleted"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at,omitempty"`
	Error      string `json:"error,omitempty"`
}

type runsListResp struct {
	Runs []runResp `json:"runs"`
}

func handleListRuns(reaper *Reaper, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		tenantID, err := uuid.Parse(q.Get("tenant_id"))
		if err != nil {
			http.Error(w, "tenant_id required", http.StatusBadRequest)
			return
		}
		limit := 20
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				if n > 100 {
					n = 100
				}
				limit = n
			}
		}

		var rows []RunEntity
		tx := reaper.DB.WithContext(r.Context()).
			Where("tenant_id = ?", tenantID).
			Order("started_at DESC").
			Limit(limit)
		if c := q.Get("category"); c != "" {
			tx = tx.Where("category = ?", c)
		}
		if t := q.Get("trigger"); t != "" {
			tx = tx.Where("\"trigger\" = ?", t)
		}
		if err := tx.Find(&rows).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		out := runsListResp{Runs: make([]runResp, 0, len(rows))}
		for _, e := range rows {
			rr := runResp{
				Service:   reaper.Service,
				RunId:     e.Id.String(),
				TenantId:  e.TenantId.String(),
				ScopeKind: e.ScopeKind,
				ScopeId:   e.ScopeId.String(),
				Category:  e.Category,
				Trigger:   e.Trigger,
				DryRun:    e.DryRun,
				Scanned:   e.Scanned,
				Deleted:   e.Deleted,
				StartedAt: e.StartedAt.UTC().Format(time.RFC3339),
			}
			if e.FinishedAt != nil {
				rr.FinishedAt = e.FinishedAt.UTC().Format(time.RFC3339)
			}
			if e.Error != nil {
				rr.Error = *e.Error
			}
			out.Runs = append(out.Runs, rr)
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// rateLimiter enforces one event per (tenant, category) per window.
type rateLimiter struct {
	mu     sync.Mutex
	window time.Duration
	last   map[string]time.Time
}

func newRateLimiter(window time.Duration) *rateLimiter {
	return &rateLimiter{window: window, last: make(map[string]time.Time)}
}

func (r *rateLimiter) allow(tenantID uuid.UUID, cat Category) bool {
	key := tenantID.String() + ":" + string(cat)
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.last[key]; ok && time.Since(t) < r.window {
		return false
	}
	r.last[key] = time.Now()
	return true
}

// silence unused-import warnings on the gorm import in builds where the
// internal handler is the only consumer.
var _ = (*gorm.DB)(nil)
