package refreshtoken

import (
	"context"
	"testing"
	"time"

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

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	raw, err := p.Create(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == "" {
		t.Error("expected non-empty token")
	}
	if len(raw) != tokenLength*2 {
		t.Errorf("expected token length %d, got %d", tokenLength*2, len(raw))
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, db *gorm.DB, p *Processor) string
		wantErr bool
	}{
		{
			name: "valid token",
			setup: func(t *testing.T, db *gorm.DB, p *Processor) string {
				raw, err := p.Create(uuid.New())
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
				return raw
			},
			wantErr: false,
		},
		{
			name: "invalid token",
			setup: func(t *testing.T, db *gorm.DB, p *Processor) string {
				return "nonexistent-token"
			},
			wantErr: true,
		},
		{
			name: "expired token",
			setup: func(t *testing.T, db *gorm.DB, p *Processor) string {
				raw, _ := generateToken()
				now := time.Now().UTC()
				db.Create(&Entity{
					Id:        uuid.New(),
					UserId:    uuid.New(),
					TokenHash: hashToken(raw),
					ExpiresAt: now.Add(-1 * time.Hour),
					Revoked:   false,
					CreatedAt: now,
					UpdatedAt: now,
				})
				return raw
			},
			wantErr: true,
		},
		{
			name: "revoked token",
			setup: func(t *testing.T, db *gorm.DB, p *Processor) string {
				raw, _ := generateToken()
				now := time.Now().UTC()
				db.Create(&Entity{
					Id:        uuid.New(),
					UserId:    uuid.New(),
					TokenHash: hashToken(raw),
					ExpiresAt: now.Add(tokenTTL),
					Revoked:   true,
					CreatedAt: now,
					UpdatedAt: now,
				})
				return raw
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			p := NewProcessor(l, context.Background(), db)

			raw := tt.setup(t, db, p)
			_, err := p.Validate(raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRotate(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	oldRaw, err := p.Create(userID)
	if err != nil {
		t.Fatalf("unexpected error creating: %v", err)
	}

	newRaw, gotUserID, err := p.Rotate(oldRaw)
	if err != nil {
		t.Fatalf("unexpected error rotating: %v", err)
	}
	if newRaw == oldRaw {
		t.Error("expected different token after rotation")
	}
	if gotUserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, gotUserID)
	}

	// Old token should now be invalid
	if _, err = p.Validate(oldRaw); err == nil {
		t.Error("expected old token to be invalid after rotation")
	}

	// New token should be valid
	if _, err = p.Validate(newRaw); err != nil {
		t.Fatalf("expected new token to be valid: %v", err)
	}
}

func TestRevokeAllForUser(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	raw1, _ := p.Create(userID)
	raw2, _ := p.Create(userID)

	if err := p.RevokeAllForUser(userID); err != nil {
		t.Fatalf("unexpected error revoking: %v", err)
	}

	if _, err := p.Validate(raw1); err == nil {
		t.Error("expected token 1 to be invalid after revoke all")
	}
	if _, err := p.Validate(raw2); err == nil {
		t.Error("expected token 2 to be invalid after revoke all")
	}
}
