// Package logging provides structured JSON logging using Logrus.
package logging

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type contextKey string

const fieldsKey contextKey = "log_fields"

// NewLogger creates a new Logrus logger configured for structured JSON output.
func NewLogger() *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.InfoLevel)
	return l
}

// WithField adds a field to the context for later inclusion in log entries.
func WithField(ctx context.Context, key string, value interface{}) context.Context {
	fields := FieldsFromContext(ctx)
	fields[key] = value
	return context.WithValue(ctx, fieldsKey, fields)
}

// FieldsFromContext retrieves log fields from the context.
func FieldsFromContext(ctx context.Context) logrus.Fields {
	if fields, ok := ctx.Value(fieldsKey).(logrus.Fields); ok {
		copied := make(logrus.Fields, len(fields))
		for k, v := range fields {
			copied[k] = v
		}
		return copied
	}
	return logrus.Fields{}
}

// WithContext returns a log entry with all fields from the context.
func WithContext(l *logrus.Logger, ctx context.Context) *logrus.Entry {
	return l.WithFields(FieldsFromContext(ctx))
}
