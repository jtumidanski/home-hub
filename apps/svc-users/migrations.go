package main

import (
	"github.com/jtumidanski/home-hub/apps/svc-users/household"
	"github.com/jtumidanski/home-hub/apps/svc-users/user"
	"gorm.io/gorm"
)

// Migration aggregates all domain migrations for the users service
// Migration order is important: households must be created before users
// due to the foreign key relationship from users to households
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// Migrate households first (no dependencies)
		if err := household.Migration()(db); err != nil {
			return err
		}

		// Migrate users second (depends on households)
		if err := user.Migration()(db); err != nil {
			return err
		}

		return nil
	}
}
