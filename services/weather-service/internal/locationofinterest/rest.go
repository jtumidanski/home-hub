package locationofinterest

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Label     *string   `json:"label,omitempty"`
	PlaceName string    `json:"placeName"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "location-of-interest" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type CreateRequest struct {
	Id        uuid.UUID `json:"-"`
	Label     *string   `json:"label,omitempty"`
	PlaceName string    `json:"placeName"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

func (r CreateRequest) GetName() string { return "location-of-interest" }
func (r CreateRequest) GetID() string   { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type UpdateRequest struct {
	Id    uuid.UUID `json:"-"`
	Label *string   `json:"label"`
}

func (r UpdateRequest) GetName() string { return "location-of-interest" }
func (r UpdateRequest) GetID() string   { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:        m.Id(),
		Label:     m.Label(),
		PlaceName: m.PlaceName(),
		Latitude:  m.Latitude(),
		Longitude: m.Longitude(),
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
	}, nil
}

func TransformSlice(models []Model) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		r, err := Transform(m)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}
