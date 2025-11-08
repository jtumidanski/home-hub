package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
)

const (
	ID = "TENANT_ID"
)

//goland:noinspection GoUnusedExportedFunction
func Creator(id uuid.UUID) ops.Provider[Model] {
	t := Model{
		id: id,
	}
	return ops.FixedProvider(t)
}

//goland:noinspection GoUnusedExportedFunction
func Create(id uuid.UUID) (Model, error) {
	return Creator(id)()
}

//goland:noinspection GoUnusedExportedFunction
func Register(id uuid.UUID) (Model, error) {
	t, err := Create(id)
	if err != nil {
		return Model{}, err
	}
	getRegistry().Add(t)
	return t, nil
}

//goland:noinspection GoUnusedExportedFunction
func AllProvider() ops.Provider[[]Model] {
	return ops.FixedProvider(getRegistry().GetAll())
}

//goland:noinspection GoUnusedExportedFunction
func ForAll(operator ops.Operator[Model]) error {
	return ops.ForEachSlice(AllProvider(), operator)
}

//goland:noinspection GoUnusedExportedFunction
func FromContext(ctx context.Context) ops.Provider[Model] {
	var ok bool
	var id uuid.UUID

	if id, ok = ctx.Value(ID).(uuid.UUID); !ok {
		return ops.ErrorProvider[Model](errors.New("unable to retrieve id from context"))
	}
	return func() (Model, error) {
		return Model{id: id}, nil
	}
}

//goland:noinspection GoUnusedExportedFunction
func MustFromContext(ctx context.Context) Model {
	t, err := FromContext(ctx)()
	if err != nil {
		panic("ctx parse err: " + err.Error())
	}
	return t
}

//goland:noinspection GoUnusedExportedFunction
func WithContext(ctx context.Context, tenant Model) context.Context {
	var wctx = ctx
	wctx = context.WithValue(wctx, ID, tenant.Id())
	return wctx
}
