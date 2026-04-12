package retention

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	sharedretention "github.com/jtumidanski/home-hub/shared/go/retention"
	"github.com/sirupsen/logrus"
)

// Errors returned by the Fanout layer.
var (
	ErrServiceUnreachable = errors.New("retention: owning service unreachable")
	ErrRateLimited        = errors.New("retention: manual purge rate limited")
)

// PurgeResult is what the owning service returned from /internal/retention/purge.
type PurgeResult struct {
	RunId   uuid.UUID
	Scanned int
	Deleted int
	DryRun  bool
}

// Fanout forwards purge requests and aggregates run lookups across the
// reaper-owning services.
type Fanout interface {
	Purge(ctx context.Context, tenantID uuid.UUID, scope sharedretention.ScopeKind, scopeID uuid.UUID, cat sharedretention.Category, dryRun bool) (PurgeResult, error)
	ListRuns(ctx context.Context, tenantID uuid.UUID, category, trigger, limit string) ([]RunRest, error)
}

// HTTPFanout is the production Fanout backed by HTTP calls. ServiceURLs maps
// service name → base URL (e.g. "productivity-service" → "http://productivity-service:8080").
type HTTPFanout struct {
	ServiceURLs map[string]string
	Token       string
	Logger      logrus.FieldLogger
	HTTP        *http.Client
}

// NewHTTPFanout constructs a Fanout. urls maps service name → base URL.
func NewHTTPFanout(urls map[string]string, token string, l logrus.FieldLogger) *HTTPFanout {
	return &HTTPFanout{
		ServiceURLs: urls,
		Token:       token,
		Logger:      l,
		HTTP:        &http.Client{Timeout: 30 * time.Second},
	}
}

// CategoryOwner returns the service name that owns a category.
func CategoryOwner(c sharedretention.Category) string {
	switch c {
	case sharedretention.CatProductivityCompletedTasks,
		sharedretention.CatProductivityDeletedTasksRestoreWindow:
		return "productivity-service"
	case sharedretention.CatRecipeDeletedRecipesRestoreWindow,
		sharedretention.CatRecipeRestorationAudit:
		return "recipe-service"
	case sharedretention.CatTrackerEntries,
		sharedretention.CatTrackerDeletedItemsRestoreWindow:
		return "tracker-service"
	case sharedretention.CatWorkoutPerformances,
		sharedretention.CatWorkoutDeletedCatalogRestoreWindow:
		return "workout-service"
	case sharedretention.CatCalendarPastEvents:
		return "calendar-service"
	case sharedretention.CatPackageArchiveWindow,
		sharedretention.CatPackageArchivedDeleteWindow:
		return "package-service"
	}
	return ""
}

// AllReaperServices returns the names of every service that exposes the
// /internal/retention/* endpoints. Used by ListRuns to fan out.
func AllReaperServices() []string {
	return []string{
		"productivity-service",
		"recipe-service",
		"tracker-service",
		"workout-service",
		"calendar-service",
		"package-service",
	}
}

type internalPurgeReq struct {
	TenantId  string `json:"tenant_id"`
	ScopeKind string `json:"scope_kind"`
	ScopeId   string `json:"scope_id"`
	Category  string `json:"category"`
	DryRun    bool   `json:"dry_run"`
}

type internalPurgeResp struct {
	RunId      string `json:"run_id"`
	Scanned    int    `json:"scanned"`
	Deleted    int    `json:"deleted"`
	DryRun     bool   `json:"dry_run"`
	DurationMs int64  `json:"duration_ms"`
}

func (f *HTTPFanout) Purge(ctx context.Context, tenantID uuid.UUID, scope sharedretention.ScopeKind, scopeID uuid.UUID, cat sharedretention.Category, dryRun bool) (PurgeResult, error) {
	owner := CategoryOwner(cat)
	base, ok := f.ServiceURLs[owner]
	if !ok {
		return PurgeResult{}, ErrServiceUnreachable
	}

	body, _ := json.Marshal(internalPurgeReq{
		TenantId:  tenantID.String(),
		ScopeKind: string(scope),
		ScopeId:   scopeID.String(),
		Category:  string(cat),
		DryRun:    dryRun,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/internal/retention/purge", bytes.NewReader(body))
	if err != nil {
		return PurgeResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if f.Token != "" {
		req.Header.Set("Authorization", "Bearer "+f.Token)
	}
	resp, err := f.HTTP.Do(req)
	if err != nil {
		return PurgeResult{}, ErrServiceUnreachable
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return PurgeResult{}, ErrRateLimited
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		return PurgeResult{}, ErrServiceUnreachable
	}
	if resp.StatusCode/100 != 2 {
		buf, _ := io.ReadAll(resp.Body)
		return PurgeResult{}, fmt.Errorf("retention: purge returned %d: %s", resp.StatusCode, string(buf))
	}
	var out internalPurgeResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return PurgeResult{}, err
	}
	rid, err := uuid.Parse(out.RunId)
	if err != nil {
		return PurgeResult{}, fmt.Errorf("retention: invalid run_id %q: %w", out.RunId, err)
	}
	return PurgeResult{RunId: rid, Scanned: out.Scanned, Deleted: out.Deleted, DryRun: out.DryRun}, nil
}

type internalRunResp struct {
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

type internalRunsList struct {
	Runs []internalRunResp `json:"runs"`
}

func (f *HTTPFanout) ListRuns(ctx context.Context, tenantID uuid.UUID, category, trigger, limit string) ([]RunRest, error) {
	if limit == "" {
		limit = "20"
	}
	results := make([]RunRest, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, svc := range AllReaperServices() {
		base, ok := f.ServiceURLs[svc]
		if !ok {
			continue
		}
		wg.Add(1)
		go func(svc, base string) {
			defer wg.Done()
			url := fmt.Sprintf("%s/internal/retention/runs?tenant_id=%s&limit=%s",
				base, tenantID, limit)
			if category != "" {
				url += "&category=" + category
			}
			if trigger != "" {
				url += "&trigger=" + trigger
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return
			}
			if f.Token != "" {
				req.Header.Set("Authorization", "Bearer "+f.Token)
			}
			resp, err := f.HTTP.Do(req)
			if err != nil {
				f.Logger.WithError(err).WithField("service", svc).Warn("retention: list runs unreachable")
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode/100 != 2 {
				return
			}
			var body internalRunsList
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for _, r := range body.Runs {
				rid, err := uuid.Parse(r.RunId)
				if err != nil {
					f.Logger.WithError(err).WithField("run_id", r.RunId).Warn("retention: skipping run with invalid id")
					continue
				}
				results = append(results, RunRest{
					Id:         rid,
					Service:    r.Service,
					Category:   r.Category,
					Scope:      r.ScopeKind,
					ScopeId:    r.ScopeId,
					Trigger:    r.Trigger,
					DryRun:     r.DryRun,
					Scanned:    r.Scanned,
					Deleted:    r.Deleted,
					StartedAt:  r.StartedAt,
					FinishedAt: r.FinishedAt,
					Error:      r.Error,
				})
			}
		}(svc, base)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool { return results[i].StartedAt > results[j].StartedAt })
	return results, nil
}
