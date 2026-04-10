package weekview

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"gorm.io/gorm"
)

// cloneItems writes a batch of cloned planned items into the target week
// inside the supplied transaction. Each row is routed through
// `planneditem.Clone` so the planneditem domain remains the only writer of
// its own table.
func cloneItems(tx *gorm.DB, tenantID, userID, targetWeekID uuid.UUID, clones []planneditem.Entity) error {
	for i := range clones {
		c := clones[i]
		c.TenantId = tenantID
		c.UserId = userID
		c.WeekId = targetWeekID
		if err := planneditem.Clone(tx, &c); err != nil {
			return err
		}
	}
	return nil
}
