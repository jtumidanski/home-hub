// Package model provides core domain types including the Provider pattern
// for lazy evaluation and functional composition.
package model

// Provider is a function that returns a value and an error.
// It enables lazy evaluation and composition of data access operations.
type Provider[T any] func() (T, error)

// FixedProvider returns a Provider that always returns the given value.
func FixedProvider[T any](value T) Provider[T] {
	return func() (T, error) {
		return value, nil
	}
}

// ErrorProvider returns a Provider that always returns the given error.
func ErrorProvider[T any](err error) Provider[T] {
	return func() (T, error) {
		var zero T
		return zero, err
	}
}

// Map transforms a Provider's output using the given function.
func Map[T any, U any](fn func(T) (U, error)) func(Provider[T]) Provider[U] {
	return func(p Provider[T]) Provider[U] {
		return func() (U, error) {
			val, err := p()
			if err != nil {
				var zero U
				return zero, err
			}
			return fn(val)
		}
	}
}

// SliceMap transforms each element in a slice Provider using the given function.
func SliceMap[T any, U any](fn func(T) (U, error)) func(Provider[[]T]) Provider[[]U] {
	return func(p Provider[[]T]) Provider[[]U] {
		return func() ([]U, error) {
			vals, err := p()
			if err != nil {
				return nil, err
			}
			results := make([]U, 0, len(vals))
			for _, v := range vals {
				r, err := fn(v)
				if err != nil {
					return nil, err
				}
				results = append(results, r)
			}
			return results, nil
		}
	}
}
