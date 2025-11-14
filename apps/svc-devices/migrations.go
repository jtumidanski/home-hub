package main

import (
	"github.com/jtumidanski/home-hub/apps/svc-devices/device"
	"github.com/jtumidanski/home-hub/apps/svc-devices/device/preference"
	"gorm.io/gorm"
)

// Migration aggregates all domain migrations for the devices service
// Migration order is important:
//   1. devices (depends on households in svc-users, but GORM handles FK externally)
//   2. device_preferences (depends on devices with CASCADE delete)
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// Migrate devices first (no internal dependencies)
		if err := device.Migration()(db); err != nil {
			return err
		}

		// Migrate device_preferences second (depends on devices)
		if err := preference.Migration()(db); err != nil {
			return err
		}

		return nil
	}
}
