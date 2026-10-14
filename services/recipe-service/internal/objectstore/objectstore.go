// Package objectstore is recipe-service's thin abstraction over an
// S3-compatible object store. It is the only package that imports minio-go.
package objectstore

import (
	"context"
	"errors"
	"io"

	"github.com/sirupsen/logrus"
)

var (
	// ErrNotFound is returned by Get when the key does not exist.
	ErrNotFound = errors.New("object not found")
	// ErrUnavailable is returned when the store is disabled or unreachable.
	ErrUnavailable = errors.New("object storage unavailable")
)

// Object is a streamed object body. Callers must close Body.
type Object struct {
	Body        io.ReadCloser
	Size        int64
	ContentType string
}

type Store interface {
	Put(ctx context.Context, key string, body []byte, contentType string) error
	Get(ctx context.Context, key string) (Object, error)
	Delete(ctx context.Context, key string) error
	// Ping checks that the configured bucket exists and is reachable.
	Ping(ctx context.Context) error
}

type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
}

// New returns a minio-backed Store, or a disabled Store when no endpoint is
// configured or the client cannot be constructed. It never blocks on the
// network.
func New(l logrus.FieldLogger, cfg Config) Store {
	if cfg.Endpoint == "" {
		l.Warn("S3_ENDPOINT not set; object storage disabled, image endpoints will return 503")
		return disabled{}
	}
	s, err := newMinio(cfg)
	if err != nil {
		l.WithError(err).Error("invalid object storage configuration; object storage disabled")
		return disabled{}
	}
	return s
}
