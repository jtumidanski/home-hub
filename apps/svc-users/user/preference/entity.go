package preference

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for user preferences.
// This is separate from the domain Model to isolate persistence concerns.
type Entity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_preferences_user_id;constraint:OnDelete:CASCADE"`
	Key       string    `gorm:"type:varchar(100);not null"`
	Value     string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for the user_preferences entity
func (e Entity) TableName() string {
	return "user_preferences"
}

// BeforeCreate sets up the UUID and timestamps before creating a new preference
func (e *Entity) BeforeCreate(tx *gorm.DB) error {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	now := time.Now()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate updates the timestamp before updating a preference
func (e *Entity) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		id:        e.Id,
		userId:    e.UserId,
		key:       e.Key,
		value:     e.Value,
		createdAt: e.CreatedAt,
		updatedAt: e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:        m.id,
		UserId:    m.userId,
		Key:       m.key,
		Value:     m.value,
		CreatedAt: m.createdAt,
		UpdatedAt: m.updatedAt,
	}
}

// Migration runs the auto-migration for the user_preferences table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// Create the table
		if err := db.AutoMigrate(&Entity{}); err != nil {
			return err
		}

		// Add unique constraint on (user_id, key) to prevent duplicate preferences
		// This ensures each user can only have one value per preference key
		if db.Migrator().HasConstraint(&Entity{}, "unique_user_key") {
			return nil
		}

		return db.Migrator().CreateConstraint(&Entity{}, "unique_user_key")
	}
}
