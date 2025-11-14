package reminder

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for reminders.
// This is separate from the domain Model to isolate persistence concerns.
type Entity struct {
	Id          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name        string     `gorm:"type:varchar(255);not null"`
	Description string     `gorm:"type:text"`
	UserId      uuid.UUID  `gorm:"type:uuid;not null;index:idx_reminders_user_remind"`
	HouseholdId uuid.UUID  `gorm:"type:uuid;not null;index:idx_reminders_household_status"`
	CreatedAt   time.Time  `gorm:"not null"`
	RemindAt    time.Time  `gorm:"not null;index:idx_reminders_user_remind;index:idx_reminders_status_remind"`
	SnoozeCount int        `gorm:"not null;default:0"`
	Status      string     `gorm:"type:varchar(20);not null;default:'active';index:idx_reminders_household_status;index:idx_reminders_status_remind"`
	DismissedAt *time.Time `gorm:"type:timestamp"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

// TableName specifies the table name for the reminders entity
func (e Entity) TableName() string {
	return "reminders"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		id:          e.Id,
		name:        e.Name,
		description: e.Description,
		userId:      e.UserId,
		householdId: e.HouseholdId,
		createdAt:   e.CreatedAt,
		remindAt:    e.RemindAt,
		snoozeCount: e.SnoozeCount,
		status:      Status(e.Status),
		dismissedAt: e.DismissedAt,
		updatedAt:   e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		Name:        m.name,
		Description: m.description,
		UserId:      m.userId,
		HouseholdId: m.householdId,
		CreatedAt:   m.createdAt,
		RemindAt:    m.remindAt,
		SnoozeCount: m.snoozeCount,
		Status:      string(m.status),
		DismissedAt: m.dismissedAt,
		UpdatedAt:   m.updatedAt,
	}
}

// Migration runs the auto-migration for the reminders table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
