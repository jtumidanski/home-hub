package authflow

import (
	"context"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/externalidentity"
	authjwt "github.com/jtumidanski/home-hub/services/auth-service/internal/jwt"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/oidc"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/refreshtoken"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/user"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CallbackResult holds the result of a successful OIDC callback.
type CallbackResult struct {
	AccessToken  string
	RefreshToken string
}

// RefreshResult holds the result of a token refresh.
type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

type Processor struct {
	l      logrus.FieldLogger
	ctx    context.Context
	db     *gorm.DB
	issuer *authjwt.Issuer
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, issuer *authjwt.Issuer) *Processor {
	return &Processor{l: l, ctx: ctx, db: db, issuer: issuer}
}

// HandleCallback processes an OIDC callback: finds or creates a user, links
// the external identity, and issues access + refresh tokens.
func (p *Processor) HandleCallback(userInfo *oidc.UserInfo) (CallbackResult, error) {
	// Find or create user
	userProc := user.NewProcessor(p.l, p.ctx, p.db)
	u, err := userProc.FindOrCreate(userInfo.Email, userInfo.DisplayName, userInfo.GivenName, userInfo.FamilyName, userInfo.AvatarURL)
	if err != nil {
		return CallbackResult{}, err
	}

	// Link external identity (idempotent — skip if already linked)
	eiProc := externalidentity.NewProcessor(p.l, p.ctx, p.db)
	_, linkErr := eiProc.FindByProviderSubject("google", userInfo.Subject)()
	if linkErr != nil {
		_, err = eiProc.Create(u.Id(), "google", userInfo.Subject)
		if err != nil {
			return CallbackResult{}, err
		}
	}

	// Issue tokens — tenant/household will be zeros until account-service onboarding
	accessToken, err := p.issuer.Issue(u.Id(), [16]byte{}, [16]byte{})
	if err != nil {
		return CallbackResult{}, err
	}

	rtProc := refreshtoken.NewProcessor(p.l, p.ctx, p.db)
	rawRefresh, err := rtProc.Create(u.Id())
	if err != nil {
		return CallbackResult{}, err
	}

	return CallbackResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}

// HandleRefresh rotates a refresh token and issues a new access token.
func (p *Processor) HandleRefresh(oldRefreshToken string) (RefreshResult, error) {
	rtProc := refreshtoken.NewProcessor(p.l, p.ctx, p.db)
	newRaw, userID, err := rtProc.Rotate(oldRefreshToken)
	if err != nil {
		return RefreshResult{}, err
	}

	// Issue new access token — tenant/household zeros (frontend resolves via context endpoint)
	accessToken, err := p.issuer.Issue(userID, [16]byte{}, [16]byte{})
	if err != nil {
		return RefreshResult{}, err
	}

	return RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRaw,
	}, nil
}

// HandleLogout revokes all refresh tokens for the given user.
func (p *Processor) HandleLogout(userID uuid.UUID) error {
	rtProc := refreshtoken.NewProcessor(p.l, p.ctx, p.db)
	return rtProc.RevokeAllForUser(userID)
}
