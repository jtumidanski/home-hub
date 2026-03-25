package task

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id              uuid.UUID  `json:"-"`
	Title           string     `json:"title"`
	Notes           string     `json:"notes,omitempty"`
	Status          string     `json:"status"`
	DueOn           *string    `json:"dueOn,omitempty"`
	RolloverEnabled bool       `json:"rolloverEnabled"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
	DeletedAt       *time.Time `json:"deletedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (r RestModel) GetName() string       { return "tasks" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model) (RestModel, error) {
	var dueOn *string
	if m.DueOn() != nil {
		s := m.DueOn().Format("2006-01-02")
		dueOn = &s
	}
	return RestModel{
		Id: m.Id(), Title: m.Title(), Notes: m.Notes(), Status: m.Status(),
		DueOn: dueOn, RolloverEnabled: m.RolloverEnabled(),
		CompletedAt: m.CompletedAt(), DeletedAt: m.DeletedAt(),
		CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt(),
	}, nil
}

func TransformSlice(models []Model) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := Transform(m)
		if err != nil {
			return nil, err
		}
		result[i] = rm
	}
	return result, nil
}

type CreateRequest struct {
	Id              uuid.UUID `json:"-"`
	Title           string    `json:"title"`
	Notes           string    `json:"notes"`
	DueOn           *string   `json:"dueOn,omitempty"`
	RolloverEnabled bool      `json:"rolloverEnabled"`
}

func (r CreateRequest) GetName() string       { return "tasks" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type UpdateRequest struct {
	Id              uuid.UUID `json:"-"`
	Title           string    `json:"title"`
	Notes           string    `json:"notes"`
	Status          string    `json:"status"`
	DueOn           *string   `json:"dueOn,omitempty"`
	RolloverEnabled bool      `json:"rolloverEnabled"`
}

func (r UpdateRequest) GetName() string       { return "tasks" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }
