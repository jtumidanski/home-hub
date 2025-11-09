package user

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a user
// Note: tenant_id and sensitive data are NEVER included in responses
type RestModel struct {
	Id          string  `jsonapi:"primary,users"`
	Email       string  `jsonapi:"attr,email"`
	DisplayName string  `jsonapi:"attr,display_name"`
	HouseholdId *string `jsonapi:"attr,household_id,omitempty"`
	CreatedAt   string  `jsonapi:"attr,created_at"`
	UpdatedAt   string  `jsonapi:"attr,updated_at"`
}

// Transform converts a domain Model to a REST representation
func Transform(m Model) (RestModel, error) {
	var householdId *string
	if m.HouseholdId() != nil {
		hid := m.HouseholdId().String()
		householdId = &hid
	}

	return RestModel{
		Id:          m.Id().String(),
		Email:       m.Email(),
		DisplayName: m.DisplayName(),
		HouseholdId: householdId,
		CreatedAt:   m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// TransformSlice converts a slice of domain Models to REST representations
func TransformSlice(models []Model) ([]RestModel, error) {
	restModels := make([]RestModel, len(models))
	for i, model := range models {
		restModel, err := Transform(model)
		if err != nil {
			return nil, err
		}
		restModels[i] = restModel
	}
	return restModels, nil
}

// CreateRequestAttributes represents the attributes for creating a user
type CreateRequestAttributes struct {
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	HouseholdId *uuid.UUID `json:"household_id,omitempty"`
}

// CreateRequest represents a JSON:API request to create a user
type CreateRequest struct {
	Data struct {
		Type       string                  `json:"type"`
		Attributes CreateRequestAttributes `json:"attributes"`
	} `json:"data"`
}

// UpdateRequestAttributes represents the attributes for updating a user
type UpdateRequestAttributes struct {
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
}

// UpdateRequest represents a JSON:API request to update a user
type UpdateRequest struct {
	Data struct {
		Type       string                  `json:"type"`
		Id         string                  `json:"id"`
		Attributes UpdateRequestAttributes `json:"attributes"`
	} `json:"data"`
}

// AssociateHouseholdRequest represents a JSON:API request to associate a household
type AssociateHouseholdRequest struct {
	Data struct {
		Type string `json:"type"`
		Id   string `json:"id"`
	} `json:"data"`
}
