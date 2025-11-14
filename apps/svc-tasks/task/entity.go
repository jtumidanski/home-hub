package task

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for tasks.
// This is separate from the domain Model to isolate persistence concerns.
type Entity struct {
	Id          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserId      uuid.UUID  `gorm:"type:uuid;not null;index:idx_tasks_user_day"`
	HouseholdId uuid.UUID  `gorm:"type:uuid;not null;index:idx_tasks_household_status"`
	Day         time.Time  `gorm:"type:date;not null;index:idx_tasks_user_day"`
	Title       string     `gorm:"type:varchar(255);not null"`
	Description string     `gorm:"type:text"`
	Status      string     `gorm:"type:varchar(20);not null;default:'incomplete';index:idx_tasks_household_status"`
	CreatedAt   time.Time  `gorm:"not null"`
	CompletedAt *time.Time `gorm:"type:timestamp"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

// TableName specifies the table name for the tasks entity
func (e Entity) TableName() string {
	return "tasks"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		id:          e.Id,
		userId:      e.UserId,
		householdId: e.HouseholdId,
		day:         e.Day,
		title:       e.Title,
		description: e.Description,
		status:      Status(e.Status),
		createdAt:   e.CreatedAt,
		completedAt: e.CompletedAt,
		updatedAt:   e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		UserId:      m.userId,
		HouseholdId: m.householdId,
		Day:         m.day,
		Title:       m.title,
		Description: m.description,
		Status:      string(m.status),
		CreatedAt:   m.createdAt,
		CompletedAt: m.completedAt,
		UpdatedAt:   m.updatedAt,
	}
}

// Migration runs the auto-migration for the tasks table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
