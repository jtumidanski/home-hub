package user

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrInvalidAvatarFormat = errors.New("avatar must be dicebear:{style}:{seed} or empty")
	validAvatarPattern     = regexp.MustCompile(`^dicebear:(adventurer|bottts|fun-emoji):[a-zA-Z0-9]{1,64}$`)
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return byIDProvider(id)(p.db.WithContext(p.ctx))
}

func (p *Processor) ByEmailProvider(email string) model.Provider[Model] {
	return byEmailProvider(email)(p.db.WithContext(p.ctx))
}

func (p *Processor) ByIDsProvider(ids []uuid.UUID) model.Provider[[]Model] {
	return model.SliceMap(modelFromEntity)(getByIDs(ids)(p.db.WithContext(p.ctx)))
}

func (p *Processor) FindOrCreate(email, displayName, givenName, familyName, avatarURL string) (Model, error) {
	m, err := p.ByEmailProvider(email)()
	if err == nil {
		return m, nil
	}

	e, err := create(p.db.WithContext(p.ctx), email, displayName, givenName, familyName, avatarURL)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

// UpdateProviderAvatar updates the provider_avatar_url column without touching the user-selected avatar_url.
func (p *Processor) UpdateProviderAvatar(userID uuid.UUID, url string) error {
	return updateProviderAvatarURL(p.db.WithContext(p.ctx), userID, url)
}

// ValidateAvatarFormat checks that avatarURL is either empty or matches dicebear:{style}:{seed}.
func ValidateAvatarFormat(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}
	if !validAvatarPattern.MatchString(avatarURL) {
		return ErrInvalidAvatarFormat
	}
	return nil
}

// UpdateAvatar validates and persists a user-selected avatar.
func (p *Processor) UpdateAvatar(userID uuid.UUID, avatarURL string) (Model, error) {
	if err := ValidateAvatarFormat(avatarURL); err != nil {
		return Model{}, err
	}

	if err := updateAvatarURL(p.db.WithContext(p.ctx), userID, avatarURL); err != nil {
		return Model{}, err
	}

	return p.ByIDProvider(userID)()
}
