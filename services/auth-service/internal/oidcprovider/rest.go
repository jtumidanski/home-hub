package oidcprovider

import "github.com/google/uuid"

// RestModel is the JSON:API representation of an OIDC provider.
type RestModel struct {
	Id          uuid.UUID `json:"-"`
	DisplayName string    `json:"displayName"`
}

func (r RestModel) GetName() string { return "auth-providers" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	r.Id = parsed
	return nil
}

// Transform converts a domain Model to a RestModel.
func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:          m.Id(),
		DisplayName: m.Name(),
	}, nil
}
