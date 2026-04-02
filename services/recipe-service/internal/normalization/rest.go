package normalization

import (
	"github.com/google/uuid"
)

type RestIngredientModel struct {
	Id                    uuid.UUID  `json:"id"`
	RawName               string     `json:"rawName"`
	RawQuantity           string     `json:"rawQuantity,omitempty"`
	RawUnit               string     `json:"rawUnit,omitempty"`
	Position              int        `json:"position"`
	CanonicalIngredientId *uuid.UUID `json:"canonicalIngredientId"`
	CanonicalName         string     `json:"canonicalName,omitempty"`
	CanonicalUnit         string     `json:"canonicalUnit,omitempty"`
	CanonicalUnitFamily   string     `json:"canonicalUnitFamily,omitempty"`
	NormalizationStatus   string     `json:"normalizationStatus"`
}

func (r RestIngredientModel) GetName() string       { return "recipe-ingredients" }
func (r RestIngredientModel) GetID() string          { return r.Id.String() }
func (r *RestIngredientModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type ResolveRequest struct {
	Id                   uuid.UUID `json:"-"`
	CanonicalIngredientId string   `json:"canonicalIngredientId"`
	SaveAsAlias          bool      `json:"saveAsAlias"`
}

func (r ResolveRequest) GetName() string       { return "ingredient-resolutions" }
func (r ResolveRequest) GetID() string          { return r.Id.String() }
func (r *ResolveRequest) SetID(id string) error {
	if id == "" { return nil }
	var err error; r.Id, err = uuid.Parse(id); return err
}

func Transform(m Model) RestIngredientModel {
	rest := RestIngredientModel{
		Id:                    m.Id(),
		RawName:               m.RawName(),
		RawQuantity:           m.RawQuantity(),
		RawUnit:               m.RawUnit(),
		Position:              m.Position(),
		CanonicalIngredientId: m.CanonicalIngredientID(),
		CanonicalUnit:         m.CanonicalUnit(),
		NormalizationStatus:   string(m.NormalizationStatus()),
	}
	if m.CanonicalUnit() != "" {
		if identity, ok := LookupUnit(m.RawUnit()); ok {
			rest.CanonicalUnitFamily = identity.Family
		}
	}
	return rest
}

type RenormalizeRequest struct {
	Id uuid.UUID `json:"-"`
}

func (r RenormalizeRequest) GetName() string       { return "recipe-renormalize" }
func (r RenormalizeRequest) GetID() string          { return r.Id.String() }
func (r *RenormalizeRequest) SetID(id string) error {
	if id == "" { return nil }
	var err error; r.Id, err = uuid.Parse(id); return err
}

func TransformSlice(models []Model) []RestIngredientModel {
	if models == nil {
		return []RestIngredientModel{}
	}
	result := make([]RestIngredientModel, len(models))
	for i, m := range models {
		result[i] = Transform(m)
	}
	return result
}
