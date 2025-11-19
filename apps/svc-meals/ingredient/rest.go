package ingredient

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of an ingredient
type RestModel struct {
	Id          uuid.UUID `json:"-"`
	MealId      string    `json:"mealId"`
	RawLine     string    `json:"rawLine"`
	Quantity    *float64  `json:"quantity"`
	QuantityRaw string    `json:"quantityRaw"`
	Unit        *string   `json:"unit"`
	UnitRaw     *string   `json:"unitRaw"`
	Ingredient  string    `json:"ingredient"`
	Preparation []string  `json:"preparation"`
	Notes       []string  `json:"notes"`
	Confidence  float64   `json:"confidence"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r RestModel) GetName() string {
	return "ingredients"
}

func (r RestModel) GetID() string {
	return r.Id.String()
}

func (r *RestModel) SetID(idStr string) error {
	if idStr == "" {
		r.Id = uuid.Nil
		return nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}

// Transform converts a domain Model to a REST representation
func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:          m.Id(),
		MealId:      m.MealId().String(),
		RawLine:     m.RawLine(),
		Quantity:    m.Quantity(),
		QuantityRaw: m.QuantityRaw(),
		Unit:        m.Unit(),
		UnitRaw:     m.UnitRaw(),
		Ingredient:  m.Ingredient(),
		Preparation: m.Preparation(),
		Notes:       m.Notes(),
		Confidence:  m.Confidence(),
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
