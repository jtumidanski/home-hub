// Package database provides GORM database connection, migration support,
// and automatic tenant filtering via callbacks.
package database

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Default connection pool settings. Tuned for a small Postgres deployment with
// max_connections ~100. The math: total services × MaxOpenConns must stay
// comfortably under Postgres max_connections, leaving headroom for psql/admin
// sessions. With ~3 services × 25 = 75 max sockets, we still have ~25 for ops.
const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 25
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

// Config holds database connection configuration.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Schema   string

	// Connection pool tuning. Zero values fall back to package defaults.
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
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

	if sqlDB, err := db.DB(); err != nil {
		l.WithError(err).Warn("failed to access underlying *sql.DB; pool not tuned")
	} else {
		maxOpen := cfg.MaxOpenConns
		if maxOpen == 0 {
			maxOpen = defaultMaxOpenConns
		}
		maxIdle := cfg.MaxIdleConns
		if maxIdle == 0 {
			maxIdle = defaultMaxIdleConns
		}
		connLifetime := cfg.ConnMaxLifetime
		if connLifetime == 0 {
			connLifetime = defaultConnMaxLifetime
		}
		connIdleTime := cfg.ConnMaxIdleTime
		if connIdleTime == 0 {
			connIdleTime = defaultConnMaxIdleTime
		}
		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetConnMaxLifetime(connLifetime)
		sqlDB.SetConnMaxIdleTime(connIdleTime)
	}

	l.WithField("schema", cfg.Schema).Info("database connected")

	if cfg.Schema != "" && cfg.Schema != "public" {
		createSchema := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q", cfg.Schema)
		if err := db.Exec(createSchema).Error; err != nil {
			l.WithError(err).Fatal("failed to create schema")
		}
	}

	RegisterTenantCallbacks(l, db)

	for _, migrate := range o.migrations {
		if err := migrate(db); err != nil {
			l.WithError(err).Fatal("migration failed")
		}
	}

	return db
}
