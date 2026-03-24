package restoration

import (
	"github.com/google/uuid"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	CreatedAt string    `json:"createdAt"`
}

func (r RestModel) GetName() string       { return "task-restorations" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type CreateRequest struct {
	Id     uuid.UUID `json:"-"`
	TaskId string    `json:"taskId"`
}

func (r CreateRequest) GetName() string       { return "task-restorations" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }
