package objectstoretest

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
)

func TestFakeRoundTrip(t *testing.T) {
	s := New()
	ctx := context.Background()
	if err := s.Put(ctx, "a/b.jpg", []byte("hello"), "image/jpeg"); err != nil {
		t.Fatal(err)
	}
	obj, err := s.Get(ctx, "a/b.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer obj.Body.Close()
	got, _ := io.ReadAll(obj.Body)
	if string(got) != "hello" || obj.Size != 5 || obj.ContentType != "image/jpeg" {
		t.Fatalf("got %q size=%d ct=%q", got, obj.Size, obj.ContentType)
	}
	if err := s.Delete(ctx, "a/b.jpg"); err != nil {
		t.Fatal(err)
	}
	if s.Has("a/b.jpg") {
		t.Fatal("object still present after delete")
	}
	if _, err := s.Get(ctx, "a/b.jpg"); !errors.Is(err, objectstore.ErrNotFound) {
		t.Fatalf("Get after delete err = %v, want ErrNotFound", err)
	}
	// Deleting a missing key is a no-op, like S3/MinIO.
	if err := s.Delete(ctx, "missing"); err != nil {
		t.Fatalf("Delete missing err = %v, want nil", err)
	}
}

func TestFakeFailureInjection(t *testing.T) {
	s := New()
	ctx := context.Background()
	boom := errors.New("boom")
	s.PutHook = func(key string) error {
		if key == "bad" {
			return boom
		}
		return nil
	}
	if err := s.Put(ctx, "good", []byte("x"), "image/jpeg"); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(ctx, "bad", []byte("x"), "image/jpeg"); !errors.Is(err, boom) {
		t.Fatalf("Put bad err = %v", err)
	}
	if s.Has("bad") {
		t.Fatal("failed put must not store the object")
	}
	s.GetErr = boom
	if _, err := s.Get(ctx, "good"); !errors.Is(err, boom) {
		t.Fatalf("Get err = %v", err)
	}
	s.DeleteErr = boom
	if err := s.Delete(ctx, "good"); !errors.Is(err, boom) {
		t.Fatalf("Delete err = %v", err)
	}
	if s.PutCalls != 2 || s.GetCalls != 1 || s.DeleteCalls != 1 {
		t.Fatalf("calls put=%d get=%d delete=%d", s.PutCalls, s.GetCalls, s.DeleteCalls)
	}
}
