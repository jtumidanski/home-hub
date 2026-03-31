package export

import "github.com/google/uuid"

// RestQuantityUnit is an additional quantity+unit pair.
type RestQuantityUnit struct {
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

// RestIngredientModel is the JSON:API representation for a consolidated ingredient.
type RestIngredientModel struct {
	Id                uuid.UUID          `json:"-"`
	Name              string             `json:"name"`
	DisplayName       *string            `json:"display_name"`
	Quantity          float64            `json:"quantity"`
	Unit              string             `json:"unit"`
	UnitFamily        string             `json:"unit_family"`
	Resolved          bool               `json:"resolved"`
	ExtraQuantities   []RestQuantityUnit `json:"extra_quantities,omitempty"`
	CategoryName      *string            `json:"category_name"`
	CategorySortOrder *int               `json:"category_sort_order"`
}

func (r RestIngredientModel) GetName() string       { return "plan-ingredients" }
func (r RestIngredientModel) GetID() string          { return r.Id.String() }
func (r *RestIngredientModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func TransformIngredientSlice(ingredients []ConsolidatedIngredient) []RestIngredientModel {
	result := make([]RestIngredientModel, len(ingredients))
	for i, ci := range ingredients {
		result[i] = TransformIngredient(ci)
	}
	return result
}

func TransformIngredient(ci ConsolidatedIngredient) RestIngredientModel {
	var displayName *string
	if ci.DisplayName != "" {
		displayName = &ci.DisplayName
	}
	var categoryName *string
	var categorySortOrder *int
	if ci.CategoryName != "" {
		categoryName = &ci.CategoryName
		categorySortOrder = &ci.CategorySortOrder
	}
	rest := RestIngredientModel{
		Id:                ci.ID,
		Name:              ci.Name,
		DisplayName:       displayName,
		Quantity:          ci.Quantity,
		Unit:              ci.Unit,
		UnitFamily:        ci.UnitFamily,
		Resolved:          ci.Resolved,
		CategoryName:      categoryName,
		CategorySortOrder: categorySortOrder,
	}
	for _, eq := range ci.ExtraQuantities {
		rest.ExtraQuantities = append(rest.ExtraQuantities, RestQuantityUnit{
			Quantity: eq.Quantity,
			Unit:     eq.Unit,
		})
	}
	return rest
}
