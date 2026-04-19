package week

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/tenant"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func GetByUserAndStart(userID uuid.UUID, weekStart time.Time) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND week_start_date = ?", userID, weekStart)
	})
}

// GetMostRecentPriorWithItems is the source-of-truth lookup for the
// "copy from previous week" endpoint. It returns the most recent week (strictly
// earlier than `before`) that owns at least one planned_items row.
func GetMostRecentPriorWithItems(db *gorm.DB, userID uuid.UUID, before time.Time) (Entity, error) {
	return nearestWithItems(db, userID, before, directionPrev)
}

// GetSoonestNextWithItems mirrors GetMostRecentPriorWithItems in the forward
// direction. It returns the soonest week (strictly later than `after`) that
// owns at least one planned_items row.
func GetSoonestNextWithItems(db *gorm.DB, userID uuid.UUID, after time.Time) (Entity, error) {
	return nearestWithItems(db, userID, after, directionNext)
}

type nearestDirection int

const (
	directionPrev nearestDirection = iota
	directionNext
)

// nearestWithItems returns the adjacent populated week in either direction.
// It qualifies the `tenant_id` filter to the `weeks` table explicitly because
// both `weeks` and `planned_items` carry `tenant_id` — leaving the tenant
// callback's unqualified filter in place causes an ambiguous-column error on
// the INNER JOIN.
func nearestWithItems(db *gorm.DB, userID uuid.UUID, anchor time.Time, dir nearestDirection) (Entity, error) {
	q := db.Model(&Entity{}).
		Joins("INNER JOIN planned_items ON planned_items.week_id = weeks.id").
		Where("weeks.user_id = ?", userID).
		Group("weeks.id")

	if dir == directionPrev {
		q = q.Where("weeks.week_start_date < ?", anchor).Order("weeks.week_start_date DESC")
	} else {
		q = q.Where("weeks.week_start_date > ?", anchor).Order("weeks.week_start_date ASC")
	}

	// The automatic tenant callback emits an unqualified `tenant_id = ?`
	// predicate that collides with the JOIN. Bypass it here and inject the
	// equivalent qualified filter so cross-tenant isolation is still enforced.
	if ctx := db.Statement.Context; ctx != nil {
		if t, ok := tenant.FromContext(ctx); ok {
			q = q.WithContext(database.WithoutTenantFilter(ctx)).
				Where("weeks.tenant_id = ?", t.Id())
		}
	}

	var e Entity
	err := q.First(&e).Error
	return e, err
}
