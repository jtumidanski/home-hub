// Package database provides GORM database connection, migration support,
// and automatic tenant filtering via callbacks.
package database

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds database connection configuration.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Schema   string
}

// Option configures the database connection.
type Option func(*connectOptions)

type connectOptions struct {
	migrations []func(*gorm.DB) error
}

// SetMigrations registers migration functions to run on startup.
func SetMigrations(fns ...func(*gorm.DB) error) Option {
	return func(o *connectOptions) {
		o.migrations = append(o.migrations, fns...)
	}
}

// Connect establishes a database connection, registers tenant callbacks,
// runs migrations, and returns the configured GORM instance.
func Connect(l *logrus.Logger, cfg Config, opts ...Option) *gorm.DB {
	o := &connectOptions{}
	for _, opt := range opts {
		opt(o)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.Schema,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		l.WithError(err).Fatal("failed to connect to database")
	}

	l.WithField("schema", cfg.Schema).Info("database connected")

	RegisterTenantCallbacks(l, db)

	for _, migrate := range o.migrations {
		if err := migrate(db); err != nil {
			l.WithError(err).Fatal("migration failed")
		}
	}

	return db
}
