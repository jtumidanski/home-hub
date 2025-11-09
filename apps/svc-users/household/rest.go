package household

// RestModel represents the JSON:API representation of a household
// Note: tenant_id and sensitive data are NEVER included in responses
type RestModel struct {
	Id        string `jsonapi:"primary,households"`
	Name      string `jsonapi:"attr,name"`
	CreatedAt string `jsonapi:"attr,created_at"`
	UpdatedAt string `jsonapi:"attr,updated_at"`
}

// Transform converts a domain Model to a REST representation
func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:        m.Id().String(),
		Name:      m.Name(),
		CreatedAt: m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
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

// CreateRequestAttributes represents the attributes for creating a household
type CreateRequestAttributes struct {
	Name string `json:"name"`
}

// CreateRequest represents a JSON:API request to create a household
type CreateRequest struct {
	Data struct {
		Type       string                  `json:"type"`
		Attributes CreateRequestAttributes `json:"attributes"`
	} `json:"data"`
}

// UpdateRequestAttributes represents the attributes for updating a household
type UpdateRequestAttributes struct {
	Name *string `json:"name,omitempty"`
}

// UpdateRequest represents a JSON:API request to update a household
type UpdateRequest struct {
	Data struct {
		Type       string                  `json:"type"`
		Id         string                  `json:"id"`
		Attributes UpdateRequestAttributes `json:"attributes"`
	} `json:"data"`
}
