package objectstore

import "context"

type disabled struct{}

func (disabled) Put(context.Context, string, []byte, string) error { return ErrUnavailable }
func (disabled) Get(context.Context, string) (Object, error)       { return Object{}, ErrUnavailable }
func (disabled) Delete(context.Context, string) error              { return ErrUnavailable }
func (disabled) Ping(context.Context) error                        { return ErrUnavailable }
