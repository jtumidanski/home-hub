package requests

import (
	"context"

	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/sirupsen/logrus"
)

//goland:noinspection GoUnusedExportedFunction
func Provider[A any, M any](l logrus.FieldLogger, ctx context.Context) func(r Request[A], t ops.Transformer[A, M]) ops.Provider[M] {
	return func(r Request[A], t ops.Transformer[A, M]) ops.Provider[M] {
		result, err := r(l, ctx)
		if err != nil {
			return ops.ErrorProvider[M](err)
		}
		return ops.Map[A, M](t)(ops.FixedProvider(result))
	}
}

//goland:noinspection GoUnusedExportedFunction
func SliceProvider[A any, M any](l logrus.FieldLogger, ctx context.Context) func(r Request[[]A], t ops.Transformer[A, M], filters []ops.Filter[M]) ops.Provider[[]M] {
	return func(r Request[[]A], t ops.Transformer[A, M], filters []ops.Filter[M]) ops.Provider[[]M] {
		resp, err := r(l, ctx)
		if err != nil {
			return ops.ErrorProvider[[]M](err)
		}
		sm := ops.SliceMap[A, M](t)(ops.FixedProvider(resp))()
		return ops.FilteredProvider[M](sm, filters)
	}
}
