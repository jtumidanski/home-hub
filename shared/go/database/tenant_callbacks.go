package database

import (
	"context"
	"reflect"
	"sync"

	"github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// hasTenantIDCache memoizes the result of struct introspection so the tenant
// callback does not walk reflect fields on every query. Keyed by reflect.Type.
var hasTenantIDCache sync.Map

type skipTenantFilterKey struct{}

// WithoutTenantFilter returns a context that bypasses automatic tenant filtering.
func WithoutTenantFilter(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipTenantFilterKey{}, true)
}

func shouldSkipTenantFilter(ctx context.Context) bool {
	skip, _ := ctx.Value(skipTenantFilterKey{}).(bool)
	return skip
}

// RegisterTenantCallbacks registers GORM callbacks that automatically inject
// WHERE tenant_id = ? on Query, Update, and Delete operations.
func RegisterTenantCallbacks(l *logrus.Logger, db *gorm.DB) {
	callback := func(db *gorm.DB) {
		if db.Statement.Context == nil {
			return
		}
		if shouldSkipTenantFilter(db.Statement.Context) {
			return
		}
		if !hasTenantIDField(db) {
			return
		}
		t, ok := tenant.FromContext(db.Statement.Context)
		if !ok {
			return
		}
		db.Statement.Where("tenant_id = ?", t.Id())
	}

	if err := db.Callback().Query().Before("gorm:query").Register("tenant:query", callback); err != nil {
		l.WithError(err).Warn("failed to register tenant query callback")
	}
	if err := db.Callback().Update().Before("gorm:update").Register("tenant:update", callback); err != nil {
		l.WithError(err).Warn("failed to register tenant update callback")
	}
	if err := db.Callback().Delete().Before("gorm:delete").Register("tenant:delete", callback); err != nil {
		l.WithError(err).Warn("failed to register tenant delete callback")
	}
}

func hasTenantIDField(db *gorm.DB) bool {
	if db.Statement.Model == nil && db.Statement.Dest == nil {
		return false
	}

	target := db.Statement.Model
	if target == nil {
		target = db.Statement.Dest
	}

	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}

	if cached, ok := hasTenantIDCache.Load(t); ok {
		return cached.(bool)
	}

	result := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("gorm")
		if field.Name == "TenantId" || contains(tag, "tenant_id") {
			result = true
			break
		}
	}
	hasTenantIDCache.Store(t, result)
	return result
}

func contains(tag, substr string) bool {
	return len(tag) > 0 && len(substr) > 0 && tagContains(tag, substr)
}

func tagContains(tag, col string) bool {
	for i := 0; i <= len(tag)-len(col); i++ {
		if tag[i:i+len(col)] == col {
			return true
		}
	}
	return false
}
