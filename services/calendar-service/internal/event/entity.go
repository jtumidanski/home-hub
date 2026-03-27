package event

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id              uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId        uuid.UUID `gorm:"type:uuid;not null;index:idx_events_tenant_household_time"`
	HouseholdId     uuid.UUID `gorm:"type:uuid;not null;index:idx_events_tenant_household_time"`
	ConnectionId    uuid.UUID `gorm:"type:uuid;not null;index:idx_events_connection"`
	SourceId        uuid.UUID `gorm:"type:uuid;not null"`
	UserId          uuid.UUID `gorm:"type:uuid;not null;index"`
	ExternalId      string    `gorm:"type:varchar(255);not null"`
	GoogleCalendarId string   `gorm:"type:varchar(255)"`
	Title           string    `gorm:"type:varchar(500);not null"`
	Description     string    `gorm:"type:text"`
	StartTime       time.Time `gorm:"not null;index:idx_events_tenant_household_time"`
	EndTime         time.Time `gorm:"not null;index:idx_events_tenant_household_time"`
	AllDay          bool      `gorm:"not null;default:false"`
	Location        string    `gorm:"type:varchar(500)"`
	Visibility      string    `gorm:"type:varchar(20);not null;default:default"`
	UserDisplayName string    `gorm:"type:varchar(255);not null"`
	UserColor       string    `gorm:"type:varchar(7);not null"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "calendar_events" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_events_source_external ON calendar_events (source_id, external_id)").Error
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:              m.id,
		TenantId:        m.tenantID,
		HouseholdId:     m.householdID,
		ConnectionId:    m.connectionID,
		SourceId:        m.sourceID,
		UserId:          m.userID,
		ExternalId:      m.externalID,
		GoogleCalendarId: m.googleCalendarID,
		Title:           m.title,
		Description:     m.description,
		StartTime:       m.startTime,
		EndTime:         m.endTime,
		AllDay:          m.allDay,
		Location:        m.location,
		Visibility:      m.visibility,
		UserDisplayName: m.userDisplayName,
		UserColor:       m.userColor,
		CreatedAt:       m.createdAt,
		UpdatedAt:       m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetConnectionID(e.ConnectionId).
		SetSourceID(e.SourceId).
		SetUserID(e.UserId).
		SetExternalID(e.ExternalId).
		SetGoogleCalendarID(e.GoogleCalendarId).
		SetTitle(e.Title).
		SetDescription(e.Description).
		SetStartTime(e.StartTime).
		SetEndTime(e.EndTime).
		SetAllDay(e.AllDay).
		SetLocation(e.Location).
		SetVisibility(e.Visibility).
		SetUserDisplayName(e.UserDisplayName).
		SetUserColor(e.UserColor).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
