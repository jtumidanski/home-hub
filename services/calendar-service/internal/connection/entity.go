package connection

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                 uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId           uuid.UUID  `gorm:"type:uuid;not null;index:idx_connections_tenant_household"`
	HouseholdId        uuid.UUID  `gorm:"type:uuid;not null;index:idx_connections_tenant_household"`
	UserId             uuid.UUID  `gorm:"type:uuid;not null;index:idx_connections_user"`
	Provider           string     `gorm:"type:varchar(50);not null"`
	Status             string     `gorm:"type:varchar(20);not null;default:connected"`
	Email              string     `gorm:"type:varchar(255);not null"`
	AccessToken        string     `gorm:"type:text;not null"`
	RefreshToken       string     `gorm:"type:text;not null"`
	TokenExpiry        time.Time  `gorm:"not null"`
	UserDisplayName    string     `gorm:"type:varchar(255);not null"`
	UserColor          string     `gorm:"type:varchar(7);not null"`
	WriteAccess        bool       `gorm:"not null;default:false"`
	LastSyncAt         *time.Time `gorm:""`
	LastSyncEventCount int        `gorm:"default:0"`
	CreatedAt          time.Time  `gorm:"not null"`
	UpdatedAt          time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "calendar_connections" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_connections_unique_provider ON calendar_connections (tenant_id, household_id, user_id, provider)").Error
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:                 m.id,
		TenantId:           m.tenantID,
		HouseholdId:        m.householdID,
		UserId:             m.userID,
		Provider:           m.provider,
		Status:             m.status,
		Email:              m.email,
		AccessToken:        m.accessToken,
		RefreshToken:       m.refreshToken,
		TokenExpiry:        m.tokenExpiry,
		UserDisplayName:    m.userDisplayName,
		UserColor:          m.userColor,
		WriteAccess:        m.writeAccess,
		LastSyncAt:         m.lastSyncAt,
		LastSyncEventCount: m.lastSyncEventCount,
		CreatedAt:          m.createdAt,
		UpdatedAt:          m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetUserID(e.UserId).
		SetProvider(e.Provider).
		SetStatus(e.Status).
		SetEmail(e.Email).
		SetAccessToken(e.AccessToken).
		SetRefreshToken(e.RefreshToken).
		SetTokenExpiry(e.TokenExpiry).
		SetUserDisplayName(e.UserDisplayName).
		SetUserColor(e.UserColor).
		SetWriteAccess(e.WriteAccess).
		SetLastSyncAt(e.LastSyncAt).
		SetLastSyncEventCount(e.LastSyncEventCount).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
