package recipeimage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
)

const (
	drainBatch   = 100
	drainTimeout = 10 * time.Second
)

// ObjectReaper drains recipe_image_object_deletions. It runs with a
// tenant-less context so the tenant callback does not filter the outbox.
// Concurrent drainers are safe: a repeated object delete is a no-op or
// ErrNotFound (success), and a repeated row delete affects nothing.
type ObjectReaper struct {
	l     logrus.FieldLogger
	db    *gorm.DB
	store objectstore.Store
}

func NewObjectReaper(l logrus.FieldLogger, db *gorm.DB, store objectstore.Store) *ObjectReaper {
	return &ObjectReaper{l: l, db: db, store: store}
}

// Run drains up to one batch, oldest first. Its signature matches the
// shared retention Loop callback.
func (r *ObjectReaper) Run(ctx context.Context) {
	rows, err := getPendingDeletions(drainBatch)(r.db.WithContext(ctx))
	if err != nil {
		r.l.WithError(err).Error("failed to load recipe image deletion outbox")
		return
	}
	r.drain(ctx, rows)
}

// DrainKeys attempts the deletes for specific outbox rows right after a
// commit. Failures stay in the outbox for Run to retry.
func (r *ObjectReaper) DrainKeys(ctx context.Context, ids []uuid.UUID) {
	if len(ids) == 0 {
		return
	}
	rows, err := getDeletionsByIDs(ids)(r.db.WithContext(ctx))
	if err != nil {
		r.l.WithError(err).Error("failed to load recipe image deletion outbox rows")
		return
	}
	r.drain(ctx, rows)
}

func (r *ObjectReaper) drain(ctx context.Context, rows []ObjectDeletionEntity) {
	for _, row := range rows {
		dctx, cancel := context.WithTimeout(ctx, drainTimeout)
		err := r.store.Delete(dctx, row.ObjectKey)
		cancel()
		if err != nil && !errors.Is(err, objectstore.ErrNotFound) {
			RecordStorageError("delete")
			r.l.WithError(err).WithFields(logrus.Fields{
				"object_key": row.ObjectKey, "attempts": row.Attempts + 1,
			}).Warn("recipe image object delete failed; will retry")
			if uerr := recordOutboxFailure(r.db.WithContext(ctx), row.Id, err.Error()); uerr != nil {
				r.l.WithError(uerr).Error("failed to record outbox failure")
			}
			continue
		}
		if derr := deleteOutboxRow(r.db.WithContext(ctx), row.Id); derr != nil {
			r.l.WithError(derr).WithField("object_key", row.ObjectKey).Error("failed to remove drained outbox row")
		}
	}
}
