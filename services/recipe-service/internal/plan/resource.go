package plan

import (
	"time"

	"github.com/google/uuid"
)

// RestListModel is the JSON:API list representation with item_count.
type RestListModel struct {
	Id        uuid.UUID `json:"-"`
	StartsOn  string    `json:"starts_on"`
	Name      string    `json:"name"`
	Locked    bool      `json:"locked"`
	ItemCount int64     `json:"item_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r RestListModel) GetName() string       { return "plans" }
func (r RestListModel) GetID() string          { return r.Id.String() }
func (r *RestListModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

// RestItemModel is the nested plan item in detail response.
type RestItemModel struct {
	Id                   uuid.UUID `json:"id"`
	Day                  string    `json:"day"`
	Slot                 string    `json:"slot"`
	RecipeID             uuid.UUID `json:"recipe_id"`
	RecipeTitle          string    `json:"recipe_title"`
	RecipeServings       *int      `json:"recipe_servings"`
	RecipeClassification string    `json:"recipe_classification,omitempty"`
	RecipeDeleted        bool      `json:"recipe_deleted"`
	ServingMultiplier    *float64  `json:"serving_multiplier"`
	PlannedServings      *int      `json:"planned_servings"`
	Notes                *string   `json:"notes"`
	Position             int       `json:"position"`
}

// RestDetailModel is the JSON:API detail representation with items.
type RestDetailModel struct {
	Id        uuid.UUID       `json:"-"`
	StartsOn  string          `json:"starts_on"`
	Name      string          `json:"name"`
	Locked    bool            `json:"locked"`
	CreatedBy uuid.UUID       `json:"created_by"`
	Items     []RestItemModel `json:"items"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (r RestDetailModel) GetName() string       { return "plans" }
func (r RestDetailModel) GetID() string          { return r.Id.String() }
func (r *RestDetailModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func TransformList(m Model, itemCount int64) RestListModel {
	return RestListModel{
		Id:        m.Id(),
		StartsOn:  m.StartsOn().Format("2006-01-02"),
		Name:      m.Name(),
		Locked:    m.Locked(),
		ItemCount: itemCount,
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
	}
}

func TransformDetail(m Model, items []RestItemModel) RestDetailModel {
	if items == nil {
		items = []RestItemModel{}
	}
	return RestDetailModel{
		Id:        m.Id(),
		StartsOn:  m.StartsOn().Format("2006-01-02"),
		Name:      m.Name(),
		Locked:    m.Locked(),
		CreatedBy: m.CreatedBy(),
		Items:     items,
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
	}
}

// CreateRequest is the JSON:API request body for creating a plan.
type CreateRequest struct {
	Id       uuid.UUID `json:"-"`
	StartsOn string    `json:"starts_on"`
	Name     string    `json:"name,omitempty"`
}

func (r CreateRequest) GetName() string       { return "plans" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// UpdateRequest is the JSON:API request body for updating a plan.
type UpdateRequest struct {
	Id   uuid.UUID `json:"-"`
	Name string    `json:"name"`
}

func (r UpdateRequest) GetName() string       { return "plans" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

// DuplicateRequest is the JSON:API request body for duplicating a plan.
type DuplicateRequest struct {
	Id       uuid.UUID `json:"-"`
	StartsOn string    `json:"starts_on"`
}

func (r DuplicateRequest) GetName() string       { return "plans" }
func (r DuplicateRequest) GetID() string          { return r.Id.String() }
func (r *DuplicateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
