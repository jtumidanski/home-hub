package user

import (
	"time"

	"github.com/google/uuid"
)

// RestModel is the JSON:API representation of a user.
type RestModel struct {
	Id                uuid.UUID `json:"-"`
	Email             string    `json:"email"`
	DisplayName       string    `json:"displayName"`
	GivenName         string    `json:"givenName"`
	FamilyName        string    `json:"familyName"`
	AvatarURL         string    `json:"avatarUrl"`
	ProviderAvatarURL string    `json:"providerAvatarUrl"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string   { return "users" }
func (r RestModel) GetID() string     { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	r.Id = parsed
	return nil
}

// Transform converts a domain Model to a RestModel.
// avatarUrl is the effective avatar: user-selected if set, otherwise empty
// (the frontend uses providerAvatarUrl as fallback).
func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:                m.Id(),
		Email:             m.Email(),
		DisplayName:       m.DisplayName(),
		GivenName:         m.GivenName(),
		FamilyName:        m.FamilyName(),
		AvatarURL:         m.AvatarURL(),
		ProviderAvatarURL: m.ProviderAvatarURL(),
		CreatedAt:         m.CreatedAt(),
		UpdatedAt:         m.UpdatedAt(),
	}, nil
}

// UpdateRequest is the JSON:API request body for PATCH /users/me.
type UpdateRequest struct {
	Id        uuid.UUID `json:"-"`
	AvatarURL string    `json:"avatarUrl"`
}

func (r UpdateRequest) GetName() string { return "users" }
func (r UpdateRequest) GetID() string   { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	r.Id = parsed
	return nil
}

// TransformSlice converts a slice of domain Models to RestModels.
func TransformSlice(models []Model) ([]RestModel, error) {
	restModels := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := Transform(m)
		if err != nil {
			return nil, err
		}
		restModels[i] = rm
	}
	return restModels, nil
}
