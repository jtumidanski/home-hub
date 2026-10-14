package recipeimage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// deleteReturning deletes the rows matching the given condition and returns
// them. In Postgres a concurrent second DELETE blocks on the row lock and
// then sees the winner's freshly inserted row, so no keys are lost.
func deleteReturning(tx *gorm.DB, query string, args ...interface{}) ([]Entity, error) {
	var removed []Entity
	err := tx.Clauses(clause.Returning{}).Where(query, args...).Delete(&removed).Error
	return removed, err
}

// deleteByRecipeIDReturning removes the recipe's image row (if any) and
// returns it.
func deleteByRecipeIDReturning(tx *gorm.DB, recipeID, householdID uuid.UUID) ([]Entity, error) {
	return deleteReturning(tx, "recipe_id = ? AND household_id = ?", recipeID, householdID)
}

func createImage(tx *gorm.DB, e *Entity) error {
	return tx.Create(e).Error
}

// enqueueObjectDeletions writes one outbox row per object key of the removed
// images and returns the outbox row ids.
func enqueueObjectDeletions(tx *gorm.DB, removed []Entity) ([]uuid.UUID, error) {
	if len(removed) == 0 {
		return nil, nil
	}
	now := time.Now().UTC()
	rows := make([]ObjectDeletionEntity, 0, len(removed)*2)
	for _, e := range removed {
		for _, key := range []string{e.FullKey, e.ThumbKey} {
			rows = append(rows, ObjectDeletionEntity{
				Id: uuid.New(), TenantId: e.TenantId, HouseholdId: e.HouseholdId,
				ObjectKey: key, CreatedAt: now,
			})
		}
	}
	if err := tx.Create(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.Id
	}
	return ids, nil
}

// PurgeByRecipeIDsTx deletes the image rows of hard-deleted recipes inside
// the caller's transaction and enqueues their objects for deletion. The
// caller (retention) has already scoped ids to one tenant + household. It
// returns the number of image rows deleted.
func PurgeByRecipeIDsTx(tx *gorm.DB, recipeIDs []uuid.UUID) (int, error) {
	if len(recipeIDs) == 0 {
		return 0, nil
	}
	removed, err := deleteReturning(tx, "recipe_id IN ?", recipeIDs)
	if err != nil {
		return 0, err
	}
	if _, err := enqueueObjectDeletions(tx, removed); err != nil {
		return 0, err
	}
	return len(removed), nil
}

func deleteOutboxRow(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&ObjectDeletionEntity{}).Error
}

func recordOutboxFailure(db *gorm.DB, id uuid.UUID, msg string) error {
	return db.Model(&ObjectDeletionEntity{}).Where("id = ?", id).
		Updates(map[string]interface{}{"attempts": gorm.Expr("attempts + 1"), "last_error": msg}).Error
}
