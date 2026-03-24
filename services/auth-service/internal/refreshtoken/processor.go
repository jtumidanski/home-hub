package refreshtoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	tokenLength = 32
	tokenTTL    = 7 * 24 * time.Hour
)

type Processor struct {
	l   *logrus.Logger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l *logrus.Logger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// Create generates a new refresh token for the given user.
// Returns the raw token (to be sent to the client) and any error.
func (p *Processor) Create(userID uuid.UUID) (string, error) {
	raw, err := generateToken()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	e := &Entity{
		Id:        uuid.New(),
		UserId:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: now.Add(tokenTTL),
		Revoked:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := p.db.WithContext(p.ctx).Create(e).Error; err != nil {
		return "", err
	}
	return raw, nil
}

// Validate checks a raw refresh token, returning the associated user ID.
func (p *Processor) Validate(raw string) (uuid.UUID, error) {
	hash := hashToken(raw)
	var e Entity
	err := p.db.WithContext(p.ctx).
		Where("token_hash = ? AND revoked = ? AND expires_at > ?", hash, false, time.Now().UTC()).
		First(&e).Error
	if err != nil {
		return uuid.Nil, errors.New("invalid or expired refresh token")
	}
	return e.UserId, nil
}

// Rotate validates the old token, revokes it, and issues a new one.
func (p *Processor) Rotate(oldRaw string) (string, uuid.UUID, error) {
	userID, err := p.Validate(oldRaw)
	if err != nil {
		return "", uuid.Nil, err
	}

	if err := p.revokeByHash(hashToken(oldRaw)); err != nil {
		return "", uuid.Nil, err
	}

	newRaw, err := p.Create(userID)
	if err != nil {
		return "", uuid.Nil, err
	}
	return newRaw, userID, nil
}

// RevokeAllForUser revokes all refresh tokens for a user.
func (p *Processor) RevokeAllForUser(userID uuid.UUID) error {
	return p.db.WithContext(p.ctx).
		Model(&Entity{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}

func (p *Processor) revokeByHash(hash string) error {
	return p.db.WithContext(p.ctx).
		Model(&Entity{}).
		Where("token_hash = ?", hash).
		Update("revoked", true).Error
}

func generateToken() (string, error) {
	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
