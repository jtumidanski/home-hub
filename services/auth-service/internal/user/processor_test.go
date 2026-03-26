package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestFindOrCreate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, p *Processor)
		email       string
		displayName string
		wantNew     bool
	}{
		{
			name:        "creates new user",
			email:       "new@example.com",
			displayName: "New User",
			wantNew:     true,
		},
		{
			name: "finds existing user",
			setup: func(t *testing.T, p *Processor) {
				_, err := p.FindOrCreate("existing@example.com", "Original Name", "Original", "Name", "")
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
			},
			email:       "existing@example.com",
			displayName: "Different Name",
			wantNew:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			p := NewProcessor(l, context.Background(), db)

			if tt.setup != nil {
				tt.setup(t, p)
			}

			m, err := p.FindOrCreate(tt.email, tt.displayName, "Given", "Family", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Email() != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, m.Email())
			}
			if m.Id() == uuid.Nil {
				t.Error("expected non-nil UUID")
			}
			if !tt.wantNew && m.DisplayName() != "Original Name" {
				t.Errorf("expected original display name, got %s", m.DisplayName())
			}
		})
	}
}

func TestValidateAvatarFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string is valid", "", false},
		{"valid adventurer", "dicebear:adventurer:seed123", false},
		{"valid bottts", "dicebear:bottts:abc", false},
		{"valid fun-emoji", "dicebear:fun-emoji:xyz789", false},
		{"invalid style", "dicebear:invalid:seed", true},
		{"arbitrary URL rejected", "https://evil.com/avatar.png", true},
		{"missing seed", "dicebear:adventurer:", true},
		{"too long seed", "dicebear:adventurer:" + string(make([]byte, 65)), true},
		{"special chars in seed", "dicebear:adventurer:seed!@#", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAvatarFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAvatarFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestUpdateAvatar(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, err := p.FindOrCreate("avatar@example.com", "Avatar User", "Avatar", "User", "https://google.com/photo.jpg")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("sets dicebear avatar", func(t *testing.T) {
		updated, err := p.UpdateAvatar(m.Id(), "dicebear:adventurer:myseed")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.AvatarURL() != "dicebear:adventurer:myseed" {
			t.Errorf("expected dicebear avatar, got %s", updated.AvatarURL())
		}
		if updated.ProviderAvatarURL() != "https://google.com/photo.jpg" {
			t.Errorf("expected provider avatar unchanged, got %s", updated.ProviderAvatarURL())
		}
	})

	t.Run("clears avatar with empty string", func(t *testing.T) {
		updated, err := p.UpdateAvatar(m.Id(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.AvatarURL() != "" {
			t.Errorf("expected empty avatar, got %s", updated.AvatarURL())
		}
	})

	t.Run("rejects invalid format", func(t *testing.T) {
		_, err := p.UpdateAvatar(m.Id(), "https://evil.com/avatar.png")
		if err == nil {
			t.Error("expected error for invalid avatar format")
		}
	})
}

func TestUpdateProviderAvatar(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, err := p.FindOrCreate("provider@example.com", "Provider User", "Provider", "User", "https://old.com/photo.jpg")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Set a user-selected avatar first
	_, err = p.UpdateAvatar(m.Id(), "dicebear:bottts:test")
	if err != nil {
		t.Fatalf("setup avatar: %v", err)
	}

	// Update provider avatar
	err = p.UpdateProviderAvatar(m.Id(), "https://new.com/photo.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user-selected avatar is NOT overwritten
	updated, err := p.ByIDProvider(m.Id())()
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if updated.AvatarURL() != "dicebear:bottts:test" {
		t.Errorf("expected user avatar unchanged, got %s", updated.AvatarURL())
	}
	if updated.ProviderAvatarURL() != "https://new.com/photo.jpg" {
		t.Errorf("expected updated provider avatar, got %s", updated.ProviderAvatarURL())
	}
}

func TestByIDProvider(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, p *Processor) uuid.UUID
		wantErr bool
	}{
		{
			name: "found",
			setup: func(t *testing.T, p *Processor) uuid.UUID {
				m, err := p.FindOrCreate("found@example.com", "Found User", "Found", "User", "")
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
				return m.Id()
			},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(t *testing.T, p *Processor) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			p := NewProcessor(l, context.Background(), db)

			id := tt.setup(t, p)
			_, err := p.ByIDProvider(id)()
			if (err != nil) != tt.wantErr {
				t.Errorf("ByIDProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
