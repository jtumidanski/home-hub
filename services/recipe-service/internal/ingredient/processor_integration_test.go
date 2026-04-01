package ingredient

import (
	"context"
	"testing"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
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
	if err := db.AutoMigrate(&Entity{}, &AliasEntity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestProcessorIntegration(t *testing.T) {
	tests := []struct {
		name        string
		createName  string
		displayName string
		unitFamily  string
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "create and retrieve ingredient",
			createName:  "flour",
			displayName: "All-Purpose Flour",
			unitFamily:  "weight",
		},
		{
			name:      "empty name returns error",
			wantErr:   true,
			wantErrIs: ErrNameRequired,
		},
		{
			name:       "invalid unit family returns error",
			createName: "sugar",
			unitFamily: "invalid",
			wantErr:    true,
			wantErrIs:  ErrInvalidUnitFamily,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)

			m, err := proc.Create(uuid.New(), tt.createName, tt.displayName, tt.unitFamily, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && err.Error() != tt.wantErrIs.Error() {
					t.Errorf("expected error %q, got %q", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := proc.Get(m.Id())
			if err != nil {
				t.Fatalf("failed to get ingredient: %v", err)
			}
			if got.Name() != tt.createName {
				t.Errorf("Name() = %q, want %q", got.Name(), tt.createName)
			}
		})
	}
}
