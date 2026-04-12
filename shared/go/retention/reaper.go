package retention

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Scope identifies a single (tenant, household-or-user) tuple a reaper acts
// on. ScopeId is the household_id or user_id depending on Kind.
type Scope struct {
	TenantId uuid.UUID
	Kind     ScopeKind
	ScopeId  uuid.UUID
}

// ReapResult is what a CategoryHandler returns from one Reap call.
type ReapResult struct {
	Scanned int
	Deleted int
}

// CategoryHandler is the per-(service, category) work unit. Each reaper-owning
// service registers one handler per category it owns.
//
//   - Category() identifies which retention rule this handler implements.
//   - DiscoverScopes() returns every (tenant, scope) tuple that has data this
//     category cares about. The reaper uses this to drive policy lookups.
//   - Reap() runs the actual delete inside the supplied tx. The handler MUST
//     respect dryRun (do all the work, return counts, and rely on the caller
//     to roll back the tx).
type CategoryHandler interface {
	Category() Category
	DiscoverScopes(ctx context.Context, db *gorm.DB) ([]Scope, error)
	Reap(ctx context.Context, tx *gorm.DB, scope Scope, retentionDays int, dryRun bool) (ReapResult, error)
}

// Reaper orchestrates a set of CategoryHandlers against a single service DB.
type Reaper struct {
	Service  string
	DB       *gorm.DB
	Policy   *PolicyClient
	Metrics  *Metrics
	Logger   logrus.FieldLogger
	Handlers []CategoryHandler
}

// New creates a Reaper.
func New(service string, db *gorm.DB, pc *PolicyClient, m *Metrics, l logrus.FieldLogger, handlers ...CategoryHandler) *Reaper {
	return &Reaper{
		Service:  service,
		DB:       db,
		Policy:   pc,
		Metrics:  m,
		Logger:   l,
		Handlers: handlers,
	}
}

// RunTick executes one full reaper pass: every handler × every scope it
// discovers, with per-scope advisory locking and audit row writing.
// Failures on one (tenant, category) are logged and recorded but never abort
// the rest of the run.
func (r *Reaper) RunTick(ctx context.Context) {
	noTenantCtx := database.WithoutTenantFilter(ctx)
	for _, h := range r.Handlers {
		select {
		case <-ctx.Done():
			return
		default:
		}
		r.runHandler(noTenantCtx, h)
	}
}

func (r *Reaper) runHandler(ctx context.Context, h CategoryHandler) {
	cat := h.Category()
	scopes, err := h.DiscoverScopes(ctx, r.DB.WithContext(ctx))
	if err != nil {
		r.Logger.WithError(err).WithField("category", cat).Warn("retention: discover scopes failed")
		return
	}
	for _, s := range scopes {
		select {
		case <-ctx.Done():
			return
		default:
		}
		r.RunOne(ctx, h, s, TriggerScheduled, false)
	}
}

// RunOne runs a single handler invocation against a specific scope. It
// performs policy lookup, advisory locking, the reap (or rollback for dry
// run), and audit row write. It is exported so the manual /internal/retention/purge
// endpoint can call it directly.
func (r *Reaper) RunOne(ctx context.Context, h CategoryHandler, scope Scope, trigger Trigger, dryRun bool) RunRecord {
	cat := h.Category()
	started := time.Now().UTC()
	rec := RunRecord{
		Id:        uuid.New(),
		TenantId:  scope.TenantId,
		ScopeKind: scope.Kind,
		ScopeId:   scope.ScopeId,
		Category:  cat,
		Trigger:   trigger,
		DryRun:    dryRun,
		StartedAt: started,
	}

	policy, err := r.Policy.GetPolicy(ctx, scope.TenantId, scope.Kind, scope.ScopeId, cat)
	if err != nil {
		r.Logger.WithError(err).WithFields(logrus.Fields{
			"tenant_id": scope.TenantId, "category": cat,
		}).Warn("retention: skipping — policy unavailable")
		fin := time.Now().UTC()
		rec.FinishedAt = &fin
		rec.Error = err.Error()
		_ = WriteRun(ctx, r.DB, rec)
		r.Metrics.ObserveRun(cat, 0, 0, 0, true)
		return rec
	}

	var result ReapResult
	var reapErr error

	txErr := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acquired, err := TryAdvisoryLock(tx, scope.TenantId, cat)
		if err != nil {
			return err
		}
		if !acquired {
			return errLockBusy
		}
		result, reapErr = h.Reap(ctx, tx, scope, policy.Days, dryRun)
		if reapErr != nil {
			return reapErr
		}
		if dryRun {
			return errDryRunRollback
		}
		return nil
	})

	failed := false
	if txErr != nil && !errors.Is(txErr, errDryRunRollback) {
		failed = true
		if errors.Is(txErr, errLockBusy) {
			r.Logger.WithFields(logrus.Fields{
				"tenant_id": scope.TenantId, "category": cat,
			}).Debug("retention: another reaper holds the lock")
		} else {
			r.Logger.WithError(txErr).WithFields(logrus.Fields{
				"tenant_id": scope.TenantId, "category": cat,
			}).Error("retention: reap failed")
		}
		rec.Error = txErr.Error()
	}

	fin := time.Now().UTC()
	rec.FinishedAt = &fin
	rec.Scanned = result.Scanned
	rec.Deleted = result.Deleted
	_ = WriteRun(ctx, r.DB, rec)

	r.Metrics.ObserveRun(cat, result.Scanned, result.Deleted, fin.Sub(started).Seconds(), failed)
	return rec
}

// HandlerFor returns the registered handler for a category, or nil.
func (r *Reaper) HandlerFor(cat Category) CategoryHandler {
	for _, h := range r.Handlers {
		if h.Category() == cat {
			return h
		}
	}
	return nil
}

// errLockBusy is returned when another reaper holds the advisory lock.
// It is logged at debug level (expected during concurrent runs) and surfaced
// to /internal/retention/purge as HTTP 409.
var errLockBusy = errors.New("retention: advisory lock busy")

// errDryRunRollback is the sentinel returned from inside a Transaction to
// force GORM to roll back without surfacing as a real error.
var errDryRunRollback = errors.New("retention: dry-run rollback")

// IsLockBusy reports whether err was caused by an advisory-lock conflict.
func IsLockBusy(err error) bool { return errors.Is(err, errLockBusy) }
