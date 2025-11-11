package main

import (
	"github.com/jtumidanski/home-hub/apps/svc-users/household"
	"github.com/jtumidanski/home-hub/apps/svc-users/user"
	"github.com/jtumidanski/home-hub/apps/svc-users/user/role"
	"gorm.io/gorm"
)

// Migration aggregates all domain migrations for the users service
// Migration order is important:
//   1. households (no dependencies)
//   2. users (depends on households)
//   3. user_roles (depends on users)
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

		// Migrate user_roles third (depends on users)
		if err := role.Migration()(db); err != nil {
			return err
		}

		return nil
	}
}
