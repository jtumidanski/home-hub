package connection

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/crypto"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/oauthstate"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound          = errors.New("connection not found")
	ErrAlreadyExists     = errors.New("user already has a connection for this provider in this household")
	ErrSyncRateLimited   = errors.New("manual sync rate limited")
	ErrNotOwner          = errors.New("connection does not belong to this user")
)

const manualSyncCooldown = 5 * time.Minute

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByUserAndHousehold(userID, householdID uuid.UUID) ([]Model, error) {
	entities, err := model.SliceMap(Make)(getByUserAndHousehold(userID, householdID)(p.db.WithContext(p.ctx)))()
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (p *Processor) AllConnected() ([]Model, error) {
	return model.SliceMap(Make)(getAllConnected()(p.noTenantDB()))()
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, provider, email, encAccessToken, encRefreshToken, displayName string, tokenExpiry time.Time) (Model, error) {
	count, err := countByHousehold(p.db.WithContext(p.ctx), householdID)
	if err != nil {
		return Model{}, err
	}
	color := UserColors[int(count)%len(UserColors)]

	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, userID, provider, email, encAccessToken, encRefreshToken, displayName, color, tokenExpiry)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) UpdateStatus(id uuid.UUID, status string) error {
	return updateStatus(p.noTenantDB(), id, status)
}

func (p *Processor) UpdateTokens(id uuid.UUID, encAccessToken string, tokenExpiry time.Time) error {
	return updateTokens(p.noTenantDB(), id, encAccessToken, tokenExpiry)
}

func (p *Processor) UpdateSyncInfo(id uuid.UUID, eventCount int) error {
	return updateSyncInfo(p.noTenantDB(), id, eventCount)
}

func (p *Processor) Delete(id uuid.UUID) error {
	return deleteByID(p.db.WithContext(p.ctx), id)
}

func (p *Processor) ByUserAndProvider(userID uuid.UUID, provider string) (Model, error) {
	return model.Map(Make)(getByUserAndProvider(userID, provider)(p.noTenantDB()))()
}

func (p *Processor) UpdateTokensAndWriteAccess(id uuid.UUID, encAccessToken, encRefreshToken string, tokenExpiry time.Time, writeAccess bool) error {
	return updateTokensAndWriteAccess(p.noTenantDB(), id, encAccessToken, encRefreshToken, tokenExpiry, writeAccess)
}

func (p *Processor) GetOrRefreshAccessToken(conn Model, gcClient *googlecal.Client, enc *crypto.Encryptor) (string, error) {
	accessToken, err := enc.Decrypt(conn.AccessToken())
	if err != nil {
		return "", err
	}

	if !conn.IsTokenExpired() {
		return accessToken, nil
	}

	refreshToken, err := enc.Decrypt(conn.RefreshToken())
	if err != nil {
		return "", err
	}

	tokenResp, err := gcClient.RefreshToken(p.ctx, refreshToken)
	if err != nil {
		return "", err
	}

	encAccess, err := enc.Encrypt(tokenResp.AccessToken)
	if err != nil {
		return "", err
	}

	tokenExpiry := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	_ = p.UpdateTokens(conn.Id(), encAccess, tokenExpiry)

	return tokenResp.AccessToken, nil
}

func (p *Processor) CreateOAuthState(tenantID, householdID, userID uuid.UUID, redirectURI string, reauthorize bool) (oauthstate.Model, error) {
	return oauthstate.NewProcessor(p.l, p.ctx, p.db).Create(tenantID, householdID, userID, redirectURI, reauthorize)
}

func (p *Processor) ValidateAndConsumeOAuthState(stateID uuid.UUID) (oauthstate.Model, error) {
	return oauthstate.NewProcessor(p.l, p.ctx, p.db).ValidateAndConsume(stateID)
}

func (p *Processor) CheckManualSyncAllowed(conn Model) error {
	if conn.lastSyncAt != nil && time.Since(*conn.lastSyncAt) < manualSyncCooldown {
		return ErrSyncRateLimited
	}
	return nil
}

func (p *Processor) noTenantDB() *gorm.DB {
	return p.db.WithContext(database.WithoutTenantFilter(p.ctx))
}
