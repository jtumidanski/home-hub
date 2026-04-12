package retention

import (
	"github.com/google/uuid"
)

// CategoryView is the per-category attribute object returned by GET.
type CategoryView struct {
	Days   int    `json:"days"`
	Source string `json:"source"`
}

// PolicyScope is the household-or-user envelope inside the GET response.
type PolicyScope struct {
	Id         string                  `json:"id"`
	Categories map[string]CategoryView `json:"categories"`
}

// PolicyRest is the JSON:API resource for GET /api/v1/retention-policies.
// The id is the tenant id. Household / User are present depending on what the
// caller has access to.
type PolicyRest struct {
	Id        uuid.UUID    `json:"-"`
	Household *PolicyScope `json:"household,omitempty"`
	User      *PolicyScope `json:"user,omitempty"`
}

func (r PolicyRest) GetName() string { return "retention-policies" }
func (r PolicyRest) GetID() string   { return r.Id.String() }
func (r *PolicyRest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// PatchRequest is the body shape for PATCH /api/v1/retention-policies/...
// Categories is a sparse map: missing key = no change, present key with int =
// upsert override, present key with explicit null = delete override. JSON:API
// unmarshalers do not preserve null vs missing distinction in a generic map,
// so we use *int and rely on the JSON decoder to set nil for explicit null.
type PatchRequest struct {
	Id         uuid.UUID    `json:"-"`
	Categories map[string]*int `json:"categories"`
}

func (r PatchRequest) GetName() string { return "retention-policies" }
func (r PatchRequest) GetID() string   { return r.Id.String() }
func (r *PatchRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// PurgeRequest is the body shape for POST /api/v1/retention-policies/purge.
type PurgeRequest struct {
	Id       uuid.UUID `json:"-"`
	Category string    `json:"category"`
	Scope    string    `json:"scope"`
	DryRun   bool      `json:"dry_run"`
}

func (r PurgeRequest) GetName() string { return "retention-purges" }
func (r PurgeRequest) GetID() string   { return r.Id.String() }
func (r *PurgeRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// PurgeResponse is the body shape for the 202 from POST .../purge.
type PurgeResponse struct {
	Id       uuid.UUID `json:"-"`
	Category string    `json:"category"`
	Scope    string    `json:"scope"`
	ScopeId  string    `json:"scope_id"`
	Status   string    `json:"status"`
	Scanned  int       `json:"scanned"`
	Deleted  int       `json:"deleted"`
	DryRun   bool      `json:"dry_run"`
}

func (r PurgeResponse) GetName() string { return "retention-purges" }
func (r PurgeResponse) GetID() string   { return r.Id.String() }
func (r *PurgeResponse) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// RunRest is the JSON:API resource for GET /api/v1/retention-runs.
type RunRest struct {
	Id         uuid.UUID `json:"-"`
	Service    string    `json:"service"`
	Category   string    `json:"category"`
	Scope      string    `json:"scope"`
	ScopeId    string    `json:"scope_id"`
	Trigger    string    `json:"trigger"`
	DryRun     bool      `json:"dry_run"`
	Scanned    int       `json:"scanned"`
	Deleted    int       `json:"deleted"`
	StartedAt  string    `json:"started_at"`
	FinishedAt string    `json:"finished_at,omitempty"`
	Error      string    `json:"error,omitempty"`
}

func (r RunRest) GetName() string { return "retention-runs" }
func (r RunRest) GetID() string   { return r.Id.String() }
func (r *RunRest) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
