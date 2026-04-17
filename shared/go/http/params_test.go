package http

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newReq(t *testing.T, query string) *nethttp.Request {
	t.Helper()
	url := "/x"
	if query != "" {
		url += "?" + query
	}
	return httptest.NewRequest(nethttp.MethodGet, url, nil)
}

func TestParseDateParam_Valid(t *testing.T) {
	r := newReq(t, "date=2026-04-16")
	got, err := ParseDateParam(r, "date")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got.Location() != time.UTC {
		t.Fatalf("expected UTC, got %v", got.Location())
	}
}

func TestParseDateParam_Missing(t *testing.T) {
	r := newReq(t, "")
	_, err := ParseDateParam(r, "date")
	if err == nil {
		t.Fatal("expected error for missing parameter")
	}
}

func TestParseDateParam_Empty(t *testing.T) {
	r := newReq(t, "date=")
	_, err := ParseDateParam(r, "date")
	if err == nil {
		t.Fatal("expected error for empty parameter")
	}
}

func TestParseDateParam_MalformedFormat(t *testing.T) {
	cases := []string{
		"date=2026-13-01",
		"date=2026-04-32",
		"date=04-16-2026",
		"date=2026/04/16",
		"date=hello",
		"date=2026-4-16",
	}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			r := newReq(t, q)
			_, err := ParseDateParam(r, "date")
			if err == nil {
				t.Fatalf("expected error for %q", q)
			}
		})
	}
}

func TestParseDateParam_WhitespacePadding(t *testing.T) {
	r := newReq(t, "date=%202026-04-16%20")
	got, err := ParseDateParam(r, "date")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseDateParam_CustomName(t *testing.T) {
	r := newReq(t, "today=2026-04-16")
	got, err := ParseDateParam(r, "today")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
