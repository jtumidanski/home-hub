package plan

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_plan_week_unique,priority:1;index:idx_plan_week_tenant_household,priority:1"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_plan_week_unique,priority:2;index:idx_plan_week_tenant_household,priority:2"`
	StartsOn    time.Time `gorm:"type:date;not null;uniqueIndex:idx_plan_week_unique,priority:3"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Locked      bool      `gorm:"not null;default:false"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "plan_weeks" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetStartsOn(e.StartsOn).
		SetName(e.Name).
		SetLocked(e.Locked).
		SetCreatedBy(e.CreatedBy).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		TenantId:    m.tenantID,
		HouseholdId: m.householdID,
		StartsOn:    m.startsOn,
		Name:        m.name,
		Locked:      m.locked,
		CreatedBy:   m.createdBy,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}
