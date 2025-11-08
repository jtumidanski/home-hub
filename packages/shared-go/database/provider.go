package database

import (
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"gorm.io/gorm"
)

type EntityProvider[E any] func(db *gorm.DB) ops.Provider[E]

func Query[E any](db *gorm.DB, query interface{}) ops.Provider[E] {
	var result E
	err := db.Where(query).First(&result).Error
	if err != nil {
		return ops.ErrorProvider[E](err)
	}
	return ops.FixedProvider[E](result)
}

func SliceQuery[E any](db *gorm.DB, query interface{}) ops.Provider[[]E] {
	var results []E
	err := db.Where(query).Find(&results).Error
	if err != nil {
		return ops.ErrorProvider[[]E](err)
	}
	return ops.FixedProvider(results)
}
