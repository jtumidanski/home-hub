package ingredient

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id          uuid.UUID  `json:"-"`
	Name        string     `json:"name"`
	DisplayName string     `json:"displayName,omitempty"`
	UnitFamily  string     `json:"unitFamily,omitempty"`
	CategoryId  *uuid.UUID `json:"categoryId,omitempty"`
	AliasCount  int        `json:"aliasCount"`
	UsageCount  int        `json:"usageCount"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (r RestModel) GetName() string       { return "ingredients" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type RestDetailModel struct {
	Id          uuid.UUID        `json:"-"`
	Name        string           `json:"name"`
	DisplayName string           `json:"displayName,omitempty"`
	UnitFamily  string           `json:"unitFamily,omitempty"`
	CategoryId  *uuid.UUID       `json:"categoryId,omitempty"`
	Aliases     []RestAliasModel `json:"aliases"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

func (r RestDetailModel) GetName() string       { return "ingredients" }
func (r RestDetailModel) GetID() string          { return r.Id.String() }
func (r *RestDetailModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type RestAliasModel struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// RestLookupModel is returned by the name-lookup endpoint. Only the fields
// other services actually need are exposed here, keeping the cross-service
// surface narrow.
type RestLookupModel struct {
	Id          uuid.UUID  `json:"-"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name"`
	CategoryId  *uuid.UUID `json:"category_id"`
}

func (r RestLookupModel) GetName() string       { return "ingredient-lookups" }
func (r RestLookupModel) GetID() string          { return r.Id.String() }
func (r *RestLookupModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func TransformLookup(m Model) RestLookupModel {
	return RestLookupModel{
		Id:          m.Id(),
		Name:        m.Name(),
		DisplayName: m.DisplayName(),
		CategoryId:  m.CategoryID(),
	}
}

type CreateRequest struct {
	Id          uuid.UUID `json:"-"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	UnitFamily  string    `json:"unitFamily"`
	CategoryId  *string   `json:"categoryId,omitempty"`
}

func (r CreateRequest) GetName() string       { return "ingredients" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" { return nil }
	var err error; r.Id, err = uuid.Parse(id); return err
}

type UpdateRequest struct {
	Id          uuid.UUID `json:"-"`
	DisplayName string    `json:"displayName,omitempty"`
	UnitFamily  string    `json:"unitFamily,omitempty"`
	Name        string    `json:"name,omitempty"`
	CategoryId  *string   `json:"categoryId"`
}

func (r UpdateRequest) GetName() string       { return "ingredients" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type AddAliasRequest struct {
	Id   uuid.UUID `json:"-"`
	Name string    `json:"name"`
}

func (r AddAliasRequest) GetName() string       { return "ingredient-aliases" }
func (r AddAliasRequest) GetID() string          { return r.Id.String() }
func (r *AddAliasRequest) SetID(id string) error {
	if id == "" { return nil }
	var err error; r.Id, err = uuid.Parse(id); return err
}

type ReassignRequest struct {
	Id                uuid.UUID `json:"-"`
	TargetIngredientId string   `json:"targetIngredientId"`
}

func (r ReassignRequest) GetName() string       { return "ingredient-reassignments" }
func (r ReassignRequest) GetID() string          { return r.Id.String() }
func (r *ReassignRequest) SetID(id string) error {
	if id == "" { return nil }
	var err error; r.Id, err = uuid.Parse(id); return err
}

type BulkCategorizeRequest struct {
	Id            uuid.UUID `json:"-"`
	IngredientIds []string  `json:"ingredient_ids"`
	CategoryId    string    `json:"category_id"`
}

func (r BulkCategorizeRequest) GetName() string       { return "ingredient-bulk-categorize" }
func (r BulkCategorizeRequest) GetID() string          { return r.Id.String() }
func (r *BulkCategorizeRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func TransformSlice(models []Model) []RestModel {
	result := make([]RestModel, len(models))
	for i, m := range models {
		result[i] = Transform(m, m.UsageCount())
	}
	return result
}

func Transform(m Model, usageCount int) RestModel {
	return RestModel{
		Id: m.Id(), Name: m.Name(), DisplayName: m.DisplayName(),
		UnitFamily: m.UnitFamily(), CategoryId: m.CategoryID(),
		AliasCount: m.AliasCount(), UsageCount: usageCount,
		CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt(),
	}
}

func TransformDetail(m Model) RestDetailModel {
	aliases := make([]RestAliasModel, len(m.Aliases()))
	for i, a := range m.Aliases() {
		aliases[i] = RestAliasModel{Id: a.Id(), Name: a.Name()}
	}
	return RestDetailModel{
		Id: m.Id(), Name: m.Name(), DisplayName: m.DisplayName(),
		UnitFamily: m.UnitFamily(), CategoryId: m.CategoryID(),
		Aliases: aliases,
		CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt(),
	}
}
