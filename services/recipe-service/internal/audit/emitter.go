package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Emit(l logrus.FieldLogger, db *gorm.DB, tenantID uuid.UUID, entityType string, entityID uuid.UUID, action string, actorID uuid.UUID, metadata map[string]interface{}) {
	var metaStr *string
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err != nil {
			l.WithError(err).Warn("Failed to marshal audit metadata")
		} else {
			s := string(b)
			metaStr = &s
		}
	}

	e := Entity{
		Id:         uuid.New(),
		TenantId:   tenantID,
		EntityType: entityType,
		EntityId:   entityID,
		Action:     action,
		ActorId:    actorID,
		Metadata:   metaStr,
		CreatedAt:  time.Now().UTC(),
	}
	if err := db.Create(&e).Error; err != nil {
		l.WithError(err).WithFields(logrus.Fields{
			"entity_type": entityType,
			"entity_id":   entityID,
			"action":      action,
		}).Warn("Failed to emit audit event")
	}
}
