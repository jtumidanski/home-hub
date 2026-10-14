package objectstore

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestNewWithEmptyEndpointIsDisabled(t *testing.T) {
	l, _ := test.NewNullLogger()
	s := New(l, Config{Bucket: "b"})
	ctx := context.Background()

	if err := s.Put(ctx, "k", []byte("x"), "image/jpeg"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Put err = %v, want ErrUnavailable", err)
	}
	if _, err := s.Get(ctx, "k"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Get err = %v, want ErrUnavailable", err)
	}
	if err := s.Delete(ctx, "k"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Delete err = %v, want ErrUnavailable", err)
	}
	if err := s.Ping(ctx); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Ping err = %v, want ErrUnavailable", err)
	}
}

func TestNewWithEndpointReturnsMinioStore(t *testing.T) {
	l, _ := test.NewNullLogger()
	s := New(l, Config{Endpoint: "localhost:9000", Bucket: "b", AccessKey: "a", SecretKey: "s", Region: "us-east-1"})
	if _, ok := s.(*minioStore); !ok {
		t.Fatalf("New returned %T, want *minioStore", s)
	}
}
