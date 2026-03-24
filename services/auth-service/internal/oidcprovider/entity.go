package oidcprovider

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM entity for auth.oidc_providers.
type Entity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:text;not null"`
	IssuerURL string    `gorm:"type:text;not null"`
	ClientID  string    `gorm:"type:text;not null"`
	Enabled   bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Entity) TableName() string {
	return "oidc_providers"
}

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

// Model is the immutable domain model for an OIDC provider.
type Model struct {
	id        uuid.UUID
	name      string
	issuerURL string
	clientID  string
	enabled   bool
}

func (m Model) Id() uuid.UUID    { return m.id }
func (m Model) Name() string     { return m.name }
func (m Model) IssuerURL() string { return m.issuerURL }
func (m Model) ClientID() string  { return m.clientID }
func (m Model) Enabled() bool    { return m.enabled }

func Make(e Entity) (Model, error) {
	return Model{
		id:        e.Id,
		name:      e.Name,
		issuerURL: e.IssuerURL,
		clientID:  e.ClientID,
		enabled:   e.Enabled,
	}, nil
}

// RestModel is the JSON:API representation of an OIDC provider.
type RestModel struct {
	Id          uuid.UUID `json:"-"`
	DisplayName string    `json:"displayName"`
}

func (r RestModel) GetName() string  { return "auth-providers" }
func (r RestModel) GetID() string    { return r.Id.String() }
func (r *RestModel) SetID(id string) { r.Id, _ = uuid.Parse(id) }

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:          m.Id(),
		DisplayName: m.Name(),
	}, nil
}
