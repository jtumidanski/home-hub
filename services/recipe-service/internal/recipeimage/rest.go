package recipeimage

import (
	"time"

	"github.com/google/uuid"
)

// RestModel is the JSON:API representation of a recipe image.
type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	ThumbURL  string    `json:"thumbUrl"`
	FullURL   string    `json:"fullUrl"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "recipe-images" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) RestModel {
	thumb, full := URLs(m.RecipeID(), m.Id())
	return RestModel{
		Id: m.Id(), Width: m.Width(), Height: m.Height(),
		ThumbURL: thumb, FullURL: full, UpdatedAt: m.UpdatedAt(),
	}
}
