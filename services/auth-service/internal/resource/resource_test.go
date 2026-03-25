package resource

import "testing"

func TestSanitizeRedirect(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty defaults to /app", input: "", want: "/app"},
		{name: "valid relative path", input: "/dashboard", want: "/dashboard"},
		{name: "valid nested path", input: "/app/settings", want: "/app/settings"},
		{name: "root path", input: "/", want: "/"},
		{name: "absolute URL rejected", input: "https://evil.com", want: "/app"},
		{name: "protocol-relative rejected", input: "//evil.com", want: "/app"},
		{name: "javascript scheme rejected", input: "javascript:alert(1)", want: "/app"},
		{name: "no leading slash rejected", input: "evil.com", want: "/app"},
		{name: "backslash rejected", input: "\\evil.com", want: "/app"},
		{name: "path with query string", input: "/app?tab=settings", want: "/app?tab=settings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeRedirect(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeRedirect(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
