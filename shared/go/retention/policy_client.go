package retention

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrPolicyUnavailable is returned when account-service is unreachable AND no
// cached value (even stale) exists. Reapers receiving this error MUST skip
// the affected (tenant, category) — never falling back to "0 days".
var ErrPolicyUnavailable = errors.New("retention: policy unavailable and no cache")

// Policy is the resolved retention window in days for a single category,
// scoped to one (tenant, scope). The Source field tracks where the value came
// from for audit and UI display.
type Policy struct {
	Days   int    `json:"days"`
	Source string `json:"source"`
}

// PolicyClient resolves retention policy for arbitrary scopes by combining
// compiled-in defaults with overrides fetched from account-service. Results
// are cached in-process for cacheTTL; on a cache miss + network failure the
// last cached value (even past TTL) is reused, and only when no cache entry
// exists at all does GetPolicy return ErrPolicyUnavailable.
type PolicyClient struct {
	baseURL  string
	token    string
	http     *http.Client
	cacheTTL time.Duration

	mu    sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	overrides map[Category]int
	fetchedAt time.Time
}

// NewPolicyClient constructs a client. baseURL is the account-service base
// (e.g. http://account-service:8080) and token is the shared internal token
// expected by /internal/retention-policies/overrides.
func NewPolicyClient(baseURL, token string) *PolicyClient {
	return &PolicyClient{
		baseURL:  baseURL,
		token:    token,
		http:     &http.Client{Timeout: 5 * time.Second},
		cacheTTL: 5 * time.Minute,
		cache:    make(map[string]cacheEntry),
	}
}

func (c *PolicyClient) cacheKey(tenantID uuid.UUID, scope ScopeKind, scopeID uuid.UUID) string {
	return tenantID.String() + ":" + string(scope) + ":" + scopeID.String()
}

// GetPolicy returns the resolved policy for one (tenant, scope, category).
// It always returns the merged value from defaults + overrides. If overrides
// cannot be loaded and there is no cache (fresh or stale) for the scope,
// it returns ErrPolicyUnavailable for that category.
func (c *PolicyClient) GetPolicy(ctx context.Context, tenantID uuid.UUID, scope ScopeKind, scopeID uuid.UUID, cat Category) (Policy, error) {
	overrides, fresh, err := c.loadOverrides(ctx, tenantID, scope, scopeID)
	if err != nil {
		return Policy{}, err
	}
	if days, ok := overrides[cat]; ok {
		src := string(scope)
		if !fresh {
			src += "(stale)"
		}
		return Policy{Days: days, Source: src}, nil
	}
	if d, ok := Defaults[cat]; ok {
		return Policy{Days: d, Source: "default"}, nil
	}
	return Policy{}, ErrUnknownCategory
}

// GetAllOverrides returns the override map for a scope (used by GET endpoints
// that want to return per-category source annotations). It honors the same
// cache + safety semantics as GetPolicy.
func (c *PolicyClient) GetAllOverrides(ctx context.Context, tenantID uuid.UUID, scope ScopeKind, scopeID uuid.UUID) (map[Category]int, error) {
	o, _, err := c.loadOverrides(ctx, tenantID, scope, scopeID)
	if err != nil {
		return nil, err
	}
	out := make(map[Category]int, len(o))
	for k, v := range o {
		out[k] = v
	}
	return out, nil
}

func (c *PolicyClient) loadOverrides(ctx context.Context, tenantID uuid.UUID, scope ScopeKind, scopeID uuid.UUID) (map[Category]int, bool, error) {
	key := c.cacheKey(tenantID, scope, scopeID)

	c.mu.Lock()
	entry, hasEntry := c.cache[key]
	c.mu.Unlock()

	if hasEntry && time.Since(entry.fetchedAt) < c.cacheTTL {
		return entry.overrides, true, nil
	}

	fetched, err := c.fetch(ctx, tenantID, scope, scopeID)
	if err != nil {
		if hasEntry {
			return entry.overrides, false, nil
		}
		return nil, false, ErrPolicyUnavailable
	}

	c.mu.Lock()
	c.cache[key] = cacheEntry{overrides: fetched, fetchedAt: time.Now()}
	c.mu.Unlock()
	return fetched, true, nil
}

type overridesResponse struct {
	Overrides map[string]int `json:"overrides"`
}

func (c *PolicyClient) fetch(ctx context.Context, tenantID uuid.UUID, scope ScopeKind, scopeID uuid.UUID) (map[Category]int, error) {
	url := fmt.Sprintf("%s/internal/retention-policies/overrides?tenant_id=%s&scope_kind=%s&scope_id=%s",
		c.baseURL, tenantID, scope, scopeID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("retention: overrides fetch returned %d", resp.StatusCode)
	}
	var body overridesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	out := make(map[Category]int, len(body.Overrides))
	for k, v := range body.Overrides {
		out[Category(k)] = v
	}
	return out, nil
}

// InvalidateScope removes the cache entry for a scope. Used after a PATCH so
// the next reaper tick picks up the new value immediately.
func (c *PolicyClient) InvalidateScope(tenantID uuid.UUID, scope ScopeKind, scopeID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, c.cacheKey(tenantID, scope, scopeID))
}
