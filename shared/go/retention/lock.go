package retention

import (
	"hash/fnv"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LockKey computes the bigint key used by pg_try_advisory_xact_lock for a
// (tenant, category) pair. The key is stable across processes and uses
// FNV-64 over the byte representation of the tenant ID and category string.
func LockKey(tenantID uuid.UUID, category Category) int64 {
	h := fnv.New64a()
	h.Write(tenantID[:])
	h.Write([]byte("|"))
	h.Write([]byte(category))
	return int64(h.Sum64())
}

// TryAdvisoryLock attempts to acquire a Postgres transaction-scoped advisory
// lock for the given (tenant, category). Must be called inside an open
// transaction (the lock is released automatically on commit/rollback).
//
// Returns true if the lock was acquired, false if another reaper already
// holds it. Returns an error only on database failure.
func TryAdvisoryLock(tx *gorm.DB, tenantID uuid.UUID, category Category) (bool, error) {
	key := LockKey(tenantID, category)
	var acquired bool
	row := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", key).Row()
	if err := row.Scan(&acquired); err != nil {
		return false, err
	}
	return acquired, nil
}
