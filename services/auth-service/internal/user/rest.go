package user

import (
	"time"

	"github.com/google/uuid"
)

// RestModel is the JSON:API representation of a user.
type RestModel struct {
	Id          uuid.UUID `json:"-"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	GivenName   string    `json:"givenName"`
	FamilyName  string    `json:"familyName"`
	AvatarURL   string    `json:"avatarUrl"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string   { return "users" }
func (r RestModel) GetID() string     { return r.Id.String() }
func (r *RestModel) SetID(id string)  { r.Id, _ = uuid.Parse(id) }

// Transform converts a domain Model to a RestModel.
func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:          m.Id(),
		Email:       m.Email(),
		DisplayName: m.DisplayName(),
		GivenName:   m.GivenName(),
		FamilyName:  m.FamilyName(),
		AvatarURL:   m.AvatarURL(),
		CreatedAt:   m.CreatedAt(),
		UpdatedAt:   m.UpdatedAt(),
	}, nil
}
