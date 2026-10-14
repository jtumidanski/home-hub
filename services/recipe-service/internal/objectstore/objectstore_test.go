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

	tests := []struct {
		name string
		call func() error
	}{
		{"Put", func() error { return s.Put(ctx, "k", []byte("x"), "image/jpeg") }},
		{"Get", func() error { _, err := s.Get(ctx, "k"); return err }},
		{"Delete", func() error { return s.Delete(ctx, "k") }},
		{"Ping", func() error { return s.Ping(ctx) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, ErrUnavailable) {
				t.Fatalf("err = %v, want ErrUnavailable", err)
			}
		})
	}
}

func TestNewWithEndpointReturnsMinioStore(t *testing.T) {
	l, _ := test.NewNullLogger()
	s := New(l, Config{Endpoint: "localhost:9000", Bucket: "b", AccessKey: "a", SecretKey: "s", Region: "us-east-1"})
	if _, ok := s.(*minioStore); !ok {
		t.Fatalf("New returned %T, want *minioStore", s)
	}
}
