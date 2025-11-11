package role

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a user role
type RestModel struct {
	UserId uuid.UUID `json:"-"`
	Role   string    `json:"role"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r RestModel) GetName() string {
	return "roles"
}

// GetID returns the composite ID (userId-role) for JSON:API
func (r RestModel) GetID() string {
	return r.UserId.String() + "-" + r.Role
}

// SetID is not supported for role resources as ID is computed
func (r *RestModel) SetID(idStr string) error {
	// ID is computed from UserId and Role, not settable
	return nil
}

// Transform converts a domain Model to a REST representation
func Transform(m Model) (RestModel, error) {
	return RestModel{
		UserId: m.UserId(),
		Role:   m.Role(),
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

// AddRoleRequest represents a JSON:API request to add a role to a user
type AddRoleRequest struct {
	UserId uuid.UUID `json:"-"`
	Role   string    `json:"role"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r AddRoleRequest) GetName() string {
	return "roles"
}

// GetID returns an empty string as this is a create request
func (r AddRoleRequest) GetID() string {
	return ""
}

// SetID is not used for role requests
func (r *AddRoleRequest) SetID(idStr string) error {
	return nil
}
