package exercise

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func ptrInt(v int) *int             { return &v }
func ptrFloat(v float64) *float64   { return &v }
func ptrString(v string) *string    { return &v }

func newValidStrengthBuilder(t *testing.T) *Builder {
	t.Helper()
	return NewBuilder().
		SetName("Bench Press").
		SetKind(KindStrength).
		SetWeightType(WeightTypeFree).
		SetThemeID(uuid.New()).
		SetRegionID(uuid.New())
}

func TestBuild_StrengthHappyPath(t *testing.T) {
	b := newValidStrengthBuilder(t).
		SetDefaultSets(ptrInt(3)).
		SetDefaultReps(ptrInt(10)).
		SetDefaultWeight(ptrFloat(135)).
		SetDefaultWeightUnit(ptrString("lb"))
	if _, err := b.Build(); err != nil {
		t.Fatalf("expected valid strength build, got %v", err)
	}
}

func TestBuild_RejectsCardioFieldsOnStrength(t *testing.T) {
	b := newValidStrengthBuilder(t).SetDefaultDistance(ptrFloat(1.0))
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidDefaultsShape) {
		t.Fatalf("expected ErrInvalidDefaultsShape, got %v", err)
	}
}

func TestBuild_RejectsRepsOnIsometric(t *testing.T) {
	b := NewBuilder().
		SetName("Plank").
		SetKind(KindIsometric).
		SetThemeID(uuid.New()).
		SetRegionID(uuid.New()).
		SetDefaultSets(ptrInt(3)).
		SetDefaultDurationSeconds(ptrInt(60)).
		SetDefaultReps(ptrInt(10))
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidDefaultsShape) {
		t.Fatalf("expected ErrInvalidDefaultsShape for isometric+reps, got %v", err)
	}
}

func TestBuild_CardioHappyPath(t *testing.T) {
	b := NewBuilder().
		SetName("Easy Run").
		SetKind(KindCardio).
		SetThemeID(uuid.New()).
		SetRegionID(uuid.New()).
		SetDefaultDurationSeconds(ptrInt(1800)).
		SetDefaultDistance(ptrFloat(3.1)).
		SetDefaultDistanceUnit(ptrString("mi"))
	if _, err := b.Build(); err != nil {
		t.Fatalf("expected valid cardio build, got %v", err)
	}
}

func TestBuild_CardioRejectsWeightFields(t *testing.T) {
	b := NewBuilder().
		SetName("Treadmill").
		SetKind(KindCardio).
		SetThemeID(uuid.New()).
		SetRegionID(uuid.New()).
		SetDefaultWeight(ptrFloat(10))
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidDefaultsShape) {
		t.Fatalf("expected ErrInvalidDefaultsShape for cardio+weight, got %v", err)
	}
}

func TestBuild_RejectsPrimaryInSecondary(t *testing.T) {
	primary := uuid.New()
	b := newValidStrengthBuilder(t).SetRegionID(primary).SetSecondaryRegionIDs([]uuid.UUID{primary})
	_, err := b.Build()
	if !errors.Is(err, ErrPrimaryInSecondary) {
		t.Fatalf("expected ErrPrimaryInSecondary, got %v", err)
	}
}

func TestBuild_RejectsInvalidKind(t *testing.T) {
	b := NewBuilder().SetName("Foo").SetKind("yoga").SetThemeID(uuid.New()).SetRegionID(uuid.New())
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidKind) {
		t.Fatalf("expected ErrInvalidKind, got %v", err)
	}
}

func TestBuild_RejectsInvalidWeightUnit(t *testing.T) {
	b := newValidStrengthBuilder(t).SetDefaultWeightUnit(ptrString("oz"))
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidWeightUnit) {
		t.Fatalf("expected ErrInvalidWeightUnit, got %v", err)
	}
}

func TestBuild_RejectsNegativeDefaults(t *testing.T) {
	b := newValidStrengthBuilder(t).SetDefaultSets(ptrInt(-1))
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidNumeric) {
		t.Fatalf("expected ErrInvalidNumeric, got %v", err)
	}
}

func TestBuild_RejectsNameTooLong(t *testing.T) {
	long := make([]byte, 101)
	for i := range long {
		long[i] = 'a'
	}
	b := newValidStrengthBuilder(t).SetName(string(long))
	_, err := b.Build()
	if !errors.Is(err, ErrNameTooLong) {
		t.Fatalf("expected ErrNameTooLong, got %v", err)
	}
}
