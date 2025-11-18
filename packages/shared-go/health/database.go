package health

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// DatabaseCheck checks the health of a database connection
type DatabaseCheck struct {
	db   *gorm.DB
	name string
}

// NewDatabaseCheck creates a new database health check
func NewDatabaseCheck(db *gorm.DB) Check {
	return &DatabaseCheck{
		db:   db,
		name: "database",
	}
}

// NewDatabaseCheckWithName creates a new database health check with a custom name
func NewDatabaseCheckWithName(db *gorm.DB, name string) Check {
	return &DatabaseCheck{
		db:   db,
		name: name,
	}
}

// Name returns the name of this check
func (c *DatabaseCheck) Name() string {
	return c.name
}

// Execute performs the database health check
func (c *DatabaseCheck) Execute(ctx context.Context) CheckResult {
	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Measure execution time
	responseTime, err := measureTime(func() error {
		var result int
		return c.db.WithContext(checkCtx).Raw("SELECT 1").Scan(&result).Error
	})

	result := CheckResult{
		ResponseTimeMs: responseTime,
	}

	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
	} else {
		result.Status = StatusHealthy
	}

	// Add connection pool stats as metadata
	if sqlDB, err := c.db.DB(); err == nil {
		stats := sqlDB.Stats()
		result.Metadata = map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"max_open_conns":   stats.MaxOpenConnections,
		}
	}

	return result
}
