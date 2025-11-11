package user

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a user
type RestModel struct {
	Id          uuid.UUID `json:"-"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Provider    string    `json:"provider"`
	HouseholdId *string   `json:"householdId,omitempty"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

// MeResponse represents the JSON:API representation of the current user with roles
type MeResponse struct {
	Id          uuid.UUID `json:"-"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Provider    string    `json:"provider"`
	HouseholdId *string   `json:"householdId,omitempty"`
	Roles       []string  `json:"roles"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

func (r *RestModel) GetName() string {
	return "users"
}

func (r RestModel) GetID() string {
	return r.Id.String()
}

func (r *RestModel) SetID(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}

func (m *MeResponse) GetName() string {
	return "me"
}

func (m MeResponse) GetID() string {
	return m.Id.String()
}

func (m *MeResponse) SetID(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	m.Id = id
	return nil
}

// Transform converts a domain Model to a REST representation
func Transform(m Model) (RestModel, error) {
	var householdId *string
	if m.HouseholdId() != nil {
		hid := m.HouseholdId().String()
		householdId = &hid
	}

	return RestModel{
		Id:          m.Id(),
		Email:       m.Email(),
		DisplayName: m.DisplayName(),
		Provider:    m.Provider(),
		HouseholdId: householdId,
		CreatedAt:   m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// TransformToMe converts a domain Model to a MeResponse representation with roles
func TransformToMe(m Model, roles []string) (MeResponse, error) {
	var householdId *string
	if m.HouseholdId() != nil {
		hid := m.HouseholdId().String()
		householdId = &hid
	}

	return MeResponse{
		Id:          m.Id(),
		Email:       m.Email(),
		DisplayName: m.DisplayName(),
		Provider:    m.Provider(),
		HouseholdId: householdId,
		Roles:       roles,
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

// CreateRequest represents a JSON:API request to create a user
type CreateRequest struct {
	Id          uuid.UUID  `json:"-"`
	Email       string     `json:"email"`
	DisplayName string     `json:"displayName"`
	HouseholdId *uuid.UUID `json:"householdId,omitempty"`
}

func (r *CreateRequest) GetName() string {
	return "users"
}

func (r CreateRequest) GetID() string {
	return r.Id.String()
}

func (r *CreateRequest) SetID(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}

// UpdateRequest represents a JSON:API request to update a user
type UpdateRequest struct {
	Id          uuid.UUID `json:"-"`
	Email       *string   `json:"email,omitempty"`
	DisplayName *string   `json:"displayName,omitempty"`
}

func (r *UpdateRequest) GetName() string {
	return "users"
}

func (r UpdateRequest) GetID() string {
	return r.Id.String()
}

func (r *UpdateRequest) SetID(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}

// AssociateHouseholdRequest represents a JSON:API request to associate a household
type AssociateHouseholdRequest struct {
	Id uuid.UUID `json:"-"`
}

func (r *AssociateHouseholdRequest) GetName() string {
	return "households"
}

func (r AssociateHouseholdRequest) GetID() string {
	return r.Id.String()
}

func (r *AssociateHouseholdRequest) SetID(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}
