// Package objectstoretest provides an in-memory objectstore.Store for tests.
package objectstoretest

import (
	"bytes"
	"context"
	"io"
	"sort"
	"sync"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
)

type stored struct {
	body        []byte
	contentType string
}

type Store struct {
	mu      sync.Mutex
	objects map[string]stored

	// PutHook, when set, is consulted before every Put; a non-nil return
	// fails that Put and nothing is stored.
	PutHook   func(key string) error
	GetErr    error
	DeleteErr error
	PingErr   error

	PutCalls    int
	GetCalls    int
	DeleteCalls int
}

func New() *Store { return &Store{objects: map[string]stored{}} }

func (s *Store) Put(_ context.Context, key string, body []byte, contentType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PutCalls++
	if s.PutHook != nil {
		if err := s.PutHook(key); err != nil {
			return err
		}
	}
	s.objects[key] = stored{body: append([]byte(nil), body...), contentType: contentType}
	return nil
}

func (s *Store) Get(_ context.Context, key string) (objectstore.Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetCalls++
	if s.GetErr != nil {
		return objectstore.Object{}, s.GetErr
	}
	o, ok := s.objects[key]
	if !ok {
		return objectstore.Object{}, objectstore.ErrNotFound
	}
	return objectstore.Object{
		Body:        io.NopCloser(bytes.NewReader(o.body)),
		Size:        int64(len(o.body)),
		ContentType: o.contentType,
	}, nil
}

func (s *Store) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DeleteCalls++
	if s.DeleteErr != nil {
		return s.DeleteErr
	}
	delete(s.objects, key)
	return nil
}

func (s *Store) Ping(context.Context) error { return s.PingErr }

func (s *Store) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.objects[key]
	return ok
}

func (s *Store) Keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]string, 0, len(s.objects))
	for k := range s.objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Seed stores an object without counting it as a Put call.
func (s *Store) Seed(key string, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = stored{body: append([]byte(nil), body...), contentType: "image/jpeg"}
}
