package invitation

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null;index"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null;index"`
	Email       string    `gorm:"type:text;not null;index:idx_invitations_email_status"`
	Role        string    `gorm:"type:text;not null;default:'viewer'"`
	Status      string    `gorm:"type:text;not null;default:'pending';index:idx_invitations_email_status"`
	InvitedBy   uuid.UUID `gorm:"type:uuid;not null"`
	ExpiresAt   time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "invitations" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	// Partial unique index: only one pending invitation per email per household.
	// Use raw SQL because GORM AutoMigrate cannot express partial unique indexes.
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_invitations_unique_pending
		ON invitations (household_id, email)
		WHERE status = 'pending'
	`).Error
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		TenantId:    m.tenantID,
		HouseholdId: m.householdID,
		Email:       m.email,
		Role:        m.role,
		Status:      m.status,
		InvitedBy:   m.invitedBy,
		ExpiresAt:   m.expiresAt,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetEmail(e.Email).
		SetRole(e.Role).
		SetStatus(e.Status).
		SetInvitedBy(e.InvitedBy).
		SetExpiresAt(e.ExpiresAt).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
