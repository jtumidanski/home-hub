package main

import (
	"github.com/jtumidanski/home-hub/apps/svc-meals/ingredient"
	"github.com/jtumidanski/home-hub/apps/svc-meals/meal"
	"gorm.io/gorm"
)

// Migration runs all database migrations for the meals service
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// Run meal migration
		if err := meal.Migration()(db); err != nil {
			return err
		}

		// Run ingredient migration
		if err := ingredient.Migration()(db); err != nil {
			return err
		}

		return nil
	}
}
