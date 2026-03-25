package database

import (
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

// EntityProvider is a function that takes a GORM DB and returns a model.Provider.
// This enables curried data access where query parameters are captured first,
// and the database connection is provided later.
type EntityProvider[T any] func(db *gorm.DB) model.Provider[T]

// Query returns a lazy Provider that executes a single-entity query when invoked.
func Query[T any](queryFn func(db *gorm.DB) *gorm.DB) EntityProvider[T] {
	return func(db *gorm.DB) model.Provider[T] {
		return func() (T, error) {
			var result T
			err := queryFn(db).First(&result).Error
			return result, err
		}
	}
}

// SliceQuery returns a lazy Provider that executes a multi-entity query when invoked.
func SliceQuery[T any](queryFn func(db *gorm.DB) *gorm.DB) EntityProvider[[]T] {
	return func(db *gorm.DB) model.Provider[[]T] {
		return func() ([]T, error) {
			var results []T
			err := queryFn(db).Find(&results).Error
			return results, err
		}
	}
}
