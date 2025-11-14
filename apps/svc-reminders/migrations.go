package main

import (
	"github.com/jtumidanski/home-hub/apps/svc-reminders/reminder"
	"gorm.io/gorm"
)

// Migration runs all database migrations for svc-reminders
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return reminder.Migration()(db)
	}
}
