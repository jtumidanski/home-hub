package database

import (
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

// EntityProvider is a function that takes a GORM DB and returns a model.Provider.
// This enables curried data access where query parameters are captured first,
// and the database connection is provided later.
type EntityProvider[T any] func(db *gorm.DB) model.Provider[T]

// Query executes a single-entity query and returns a Provider.
func Query[T any](queryFn func(db *gorm.DB) *gorm.DB) EntityProvider[T] {
	return func(db *gorm.DB) model.Provider[T] {
		var result T
		err := queryFn(db).First(&result).Error
		if err != nil {
			return model.ErrorProvider[T](err)
		}
		return model.FixedProvider(result)
	}
}

// SliceQuery executes a multi-entity query and returns a Provider of slices.
func SliceQuery[T any](queryFn func(db *gorm.DB) *gorm.DB) EntityProvider[[]T] {
	return func(db *gorm.DB) model.Provider[[]T] {
		var results []T
		err := queryFn(db).Find(&results).Error
		if err != nil {
			return model.ErrorProvider[[]T](err)
		}
		return model.FixedProvider(results)
	}
}
