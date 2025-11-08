package async

import (
	"context"
	"errors"
	"time"

	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
)

var ErrAwaitTimeout = errors.New("timeout")

type Config struct {
	ctx     context.Context
	timeout time.Duration
}

type Configurator func(*Config)

//goland:noinspection GoUnusedExportedFunction
func SetContext(ctx context.Context) Configurator {
	return func(config *Config) {
		config.ctx = ctx
	}
}

//goland:noinspection GoUnusedExportedFunction
func SetTimeout(timeout time.Duration) Configurator {
	return func(config *Config) {
		config.timeout = timeout
	}
}

type Provider[M any] func(ctx context.Context, rchan chan M, echan chan error)

//goland:noinspection GoUnusedExportedFunction
func SingleProvider[M any](p Provider[M]) ops.Provider[Provider[M]] {
	return ops.FixedProvider[Provider[M]](p)
}

//goland:noinspection GoUnusedExportedFunction
func FixedProvider[M any](ps []Provider[M]) ops.Provider[[]Provider[M]] {
	return ops.FixedProvider[[]Provider[M]](ps)
}

//goland:noinspection GoUnusedExportedFunction
func Await[M any](provider ops.Provider[Provider[M]], configurators ...Configurator) ops.Provider[M] {
	return ops.FirstProvider(AwaitSlice(ops.ToSliceProvider(provider), configurators...), ops.Filters[M]())
}

//goland:noinspection GoUnusedExportedFunction
func AwaitSlice[M any](provider ops.Provider[[]Provider[M]], configurators ...Configurator) ops.Provider[[]M] {
	c := &Config{ctx: context.Background(), timeout: 500 * time.Millisecond}
	for _, configurator := range configurators {
		configurator(c)
	}

	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	providers, err := provider()
	if err != nil {
		return ops.ErrorProvider[[]M](err)
	}

	resultChannels := make(chan M, len(providers))
	errChannels := make(chan error, len(providers))

	for _, provider := range providers {
		p := provider
		go func() {
			p(ctx, resultChannels, errChannels)
		}()
	}

	var results = make([]M, 0)
	for i := 0; i < len(providers); i++ {
		select {
		case result := <-resultChannels:
			results = append(results, result)
		case <-ctx.Done():
			err = ErrAwaitTimeout
		case <-errChannels:
			err = <-errChannels
		}
	}
	return func() ([]M, error) {
		return results, err
	}
}
