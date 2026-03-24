package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Processor handles user business logic.
type Processor struct {
	l   *logrus.Logger
	ctx context.Context
	db  *gorm.DB
}

// NewProcessor creates a new user Processor.
func NewProcessor(l *logrus.Logger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// ByIDProvider returns a provider for a user by ID.
func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))
}

// ByEmailProvider returns a provider for a user by email.
func (p *Processor) ByEmailProvider(email string) model.Provider[Model] {
	return model.Map(Make)(getByEmail(email)(p.db.WithContext(p.ctx)))
}

// FindOrCreate looks up a user by email. If not found, creates one.
func (p *Processor) FindOrCreate(email, displayName, givenName, familyName, avatarURL string) (Model, error) {
	m, err := p.ByEmailProvider(email)()
	if err == nil {
		return m, nil
	}

	return p.create(email, displayName, givenName, familyName, avatarURL)
}

func (p *Processor) create(email, displayName, givenName, familyName, avatarURL string) (Model, error) {
	now := time.Now().UTC()
	e := &Entity{
		Id:          uuid.New(),
		Email:       email,
		DisplayName: displayName,
		GivenName:   givenName,
		FamilyName:  familyName,
		AvatarURL:   avatarURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := p.db.WithContext(p.ctx).Create(e).Error; err != nil {
		return Model{}, err
	}
	return Make(*e)
}
