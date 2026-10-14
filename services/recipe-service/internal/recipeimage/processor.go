package recipeimage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage/imaging"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

const (
	contentTypeJPEG = "image/jpeg"
	putTimeout      = 30 * time.Second
	cleanupTimeout  = 10 * time.Second
)

// processSem bounds concurrent decodes: a 40 MP PNG decodes to ~160 MB.
var processSem = make(chan struct{}, 2)

// beforeInsertHook is a test seam that runs inside the swap transaction
// between the DELETE and the INSERT. Always nil in production.
var beforeInsertHook func(tx *gorm.DB, e *Entity) error

type Processor struct {
	l     logrus.FieldLogger
	ctx   context.Context
	db    *gorm.DB
	store objectstore.Store
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// WithStore returns a copy bound to an object store. Read-only callers
// (recipe enrichment) do not need one.
func (p *Processor) WithStore(s objectstore.Store) *Processor {
	c := *p
	c.store = s
	return &c
}

type UploadResult struct {
	Model          Model
	Replaced       bool
	OutboxIDs      []uuid.UUID
	SourceBytes    int
	ProcessingTime time.Duration
}

func (p *Processor) Upload(recipeID uuid.UUID, src []byte) (UploadResult, error) {
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return UploadResult{}, ErrNoTenant
	}

	start := time.Now()
	res, err := p.process(src)
	elapsed := time.Since(start)
	if err != nil {
		return UploadResult{}, err
	}
	processingSeconds.Observe(elapsed.Seconds())

	imageID := uuid.New()
	fullKey, thumbKey := objectKeys(t.Id(), t.HouseholdId(), recipeID, imageID)
	log := p.l.WithFields(logrus.Fields{"recipe_id": recipeID, "image_id": imageID})
	// Logged before Put so an operator can find orphans from a crash
	// between Put and commit (design §9).
	log.Info("storing recipe image renditions")

	storeStart := time.Now()
	if err := p.put(fullKey, res.Full.Bytes); err != nil {
		return UploadResult{}, err
	}
	if err := p.put(thumbKey, res.Thumb.Bytes); err != nil {
		p.deleteNow(fullKey)
		return UploadResult{}, err
	}
	storageElapsed := time.Since(storeStart)

	now := time.Now().UTC()
	e := Entity{
		Id: imageID, TenantId: t.Id(), HouseholdId: t.HouseholdId(), RecipeId: recipeID,
		FullKey: fullKey, ThumbKey: thumbKey, ContentType: contentTypeJPEG,
		Width: res.Full.Width, Height: res.Full.Height,
		FullBytes: len(res.Full.Bytes), ThumbBytes: len(res.Thumb.Bytes),
		CreatedAt: now, UpdatedAt: now,
	}
	replaced, outboxIDs, err := p.swap(&e)
	if err != nil {
		p.deleteNow(fullKey)
		p.deleteNow(thumbKey)
		return UploadResult{}, err
	}

	m, err := Make(e)
	if err != nil {
		return UploadResult{}, err
	}
	log.WithFields(logrus.Fields{
		"source_bytes": len(src), "full_bytes": e.FullBytes, "thumb_bytes": e.ThumbBytes,
		"processing_ms": elapsed.Milliseconds(), "storage_ms": storageElapsed.Milliseconds(),
		"replaced": replaced,
	}).Info("recipe image stored")
	return UploadResult{Model: m, Replaced: replaced, OutboxIDs: outboxIDs, SourceBytes: len(src), ProcessingTime: elapsed}, nil
}

func (p *Processor) process(src []byte) (imaging.Result, error) {
	select {
	case processSem <- struct{}{}:
	case <-p.ctx.Done():
		return imaging.Result{}, p.ctx.Err()
	}
	defer func() { <-processSem }()
	return imaging.Process(src)
}

func (p *Processor) put(key string, body []byte) error {
	ctx, cancel := context.WithTimeout(p.ctx, putTimeout)
	defer cancel()
	if err := p.store.Put(ctx, key, body, contentTypeJPEG); err != nil {
		RecordStorageError("put")
		p.l.WithError(err).WithField("object_key", key).Error("failed to store recipe image object")
		return fmt.Errorf("%w: %v", ErrStorage, err)
	}
	return nil
}

// deleteNow is a best-effort, direct (non-outbox) delete used only to clean
// up objects that no committed row references.
func (p *Processor) deleteNow(key string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(p.ctx), cleanupTimeout)
	defer cancel()
	if err := p.store.Delete(ctx, key); err != nil && !errors.Is(err, objectstore.ErrNotFound) {
		RecordStorageError("delete")
		p.l.WithError(err).WithField("object_key", key).Warn("failed to clean up recipe image object")
	}
}

func (p *Processor) swap(e *Entity) (bool, []uuid.UUID, error) {
	replaced, ids, err := p.swapTx(e)
	if err != nil && isUniqueViolation(err) {
		p.l.WithField("recipe_id", e.RecipeId).Warn("concurrent first image upload; retrying swap once")
		replaced, ids, err = p.swapTx(e)
	}
	return replaced, ids, err
}

func (p *Processor) swapTx(e *Entity) (replaced bool, outboxIDs []uuid.UUID, err error) {
	err = p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		e.CreatedAt = e.UpdatedAt
		removed, err := deleteByRecipeIDReturning(tx, e.RecipeId, e.HouseholdId)
		if err != nil {
			return err
		}
		if len(removed) > 0 {
			replaced = true
			e.CreatedAt = removed[0].CreatedAt
		}
		if beforeInsertHook != nil {
			if err := beforeInsertHook(tx, e); err != nil {
				return err
			}
		}
		if err := createImage(tx, e); err != nil {
			return err
		}
		outboxIDs, err = enqueueObjectDeletions(tx, removed)
		return err
	})
	if err != nil {
		return false, nil, err
	}
	return replaced, outboxIDs, nil
}

// Delete removes the recipe's image row and enqueues its objects. Deleting a
// recipe with no image succeeds with no ids.
func (p *Processor) Delete(recipeID uuid.UUID) ([]uuid.UUID, error) {
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return nil, ErrNoTenant
	}
	var ids []uuid.UUID
	err := p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		removed, err := deleteByRecipeIDReturning(tx, recipeID, t.HouseholdId())
		if err != nil {
			return err
		}
		ids, err = enqueueObjectDeletions(tx, removed)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		p.l.WithField("recipe_id", recipeID).Info("recipe image deleted")
	}
	return ids, nil
}

func (p *Processor) GetByRecipeID(recipeID uuid.UUID) (Model, error) {
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return Model{}, ErrNoTenant
	}
	e, err := getByRecipeID(recipeID, t.HouseholdId())(p.db.WithContext(p.ctx))()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Model{}, ErrImageNotFound
	}
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

// GetByRecipeIDs returns images keyed by recipe id in a single query. Empty
// input returns an empty map without hitting the database.
func (p *Processor) GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID]Model, error) {
	result := make(map[uuid.UUID]Model)
	if len(recipeIDs) == 0 {
		return result, nil
	}
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return nil, ErrNoTenant
	}
	entities, err := getByRecipeIDs(recipeIDs, t.HouseholdId())(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		result[e.RecipeId] = m
	}
	return result, nil
}

// Open streams one rendition. The caller must close the body.
func (p *Processor) Open(m Model, size Size) (objectstore.Object, error) {
	obj, err := p.store.Get(p.ctx, m.Key(size))
	if errors.Is(err, objectstore.ErrNotFound) {
		p.l.WithFields(logrus.Fields{"recipe_id": m.RecipeID(), "image_id": m.Id(), "object_key": m.Key(size)}).
			Error("recipe image row points at a missing object")
		return objectstore.Object{}, ErrObjectMissing
	}
	if err != nil {
		RecordStorageError("get")
		return objectstore.Object{}, fmt.Errorf("%w: %v", ErrStorage, err)
	}
	return obj, nil
}
