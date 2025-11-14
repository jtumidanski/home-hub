package main

import (
	"github.com/jtumidanski/home-hub/apps/svc-tasks/task"
	"gorm.io/gorm"
)

// Migration aggregates all domain migrations for the tasks service
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// Migrate tasks table
		if err := task.Migration()(db); err != nil {
			return err
		}

		return nil
	}
}
