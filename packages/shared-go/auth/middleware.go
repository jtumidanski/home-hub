package auth

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserProvider defines the interface for loading user data
type UserProvider interface {
	GetOrCreateUserByEmail(email, displayName, provider string) (uuid.UUID, error)
}

// RoleProvider defines the interface for loading user roles
type RoleProvider interface {
	GetRolesByUserId(userId uuid.UUID) ([]string, error)
}

// Middleware creates an HTTP middleware that enforces authentication
// It extracts auth headers, validates source, loads/creates user, loads roles
func Middleware(logger *logrus.Logger, db *gorm.DB, userProvider UserProvider, roleProvider RoleProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Validate request source
			if err := ValidateSource(r); err != nil {
				logger.WithError(err).Warn("Request from untrusted source")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 2. Extract auth headers
			headers, err := ExtractHeaders(r)
			if err != nil {
				logger.WithError(err).Warn("Failed to extract auth headers")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 3. Infer provider
			provider := InferProvider(headers)

			// 4. Get or create user
			userId, err := userProvider.GetOrCreateUserByEmail(headers.Email, headers.User, provider)
			if err != nil {
				logger.WithError(err).Error("Failed to get or create user")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// 5. Load user roles
			roles, err := roleProvider.GetRolesByUserId(userId)
			if err != nil {
				logger.WithError(err).Error("Failed to load user roles")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// 6. Create auth context
			authCtx := Context{
				UserId:   userId,
				Email:    headers.Email,
				Name:     headers.User,
				Provider: provider,
				Roles:    roles,
			}

			// 7. Attach to request context
			ctx := WithContext(r.Context(), authCtx)

			// Log successful authentication
			logger.WithFields(logrus.Fields{
				"user_id":  userId.String(),
				"email":    headers.Email,
				"provider": provider,
				"roles":    roles,
			}).Debug("User authenticated")

			// 8. Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalMiddleware is similar to Middleware but doesn't fail if auth headers are missing
// Useful for endpoints that have both authenticated and public access
func OptionalMiddleware(logger *logrus.Logger, db *gorm.DB, userProvider UserProvider, roleProvider RoleProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to validate source
			if err := ValidateSource(r); err != nil {
				// Not from ingress, continue without auth
				next.ServeHTTP(w, r)
				return
			}

			// Try to extract headers
			headers, err := ExtractHeaders(r)
			if err != nil {
				// No auth headers, continue without auth
				next.ServeHTTP(w, r)
				return
			}

			// From here, same as required middleware
			provider := InferProvider(headers)

			userId, err := userProvider.GetOrCreateUserByEmail(headers.Email, headers.User, provider)
			if err != nil {
				logger.WithError(err).Warn("Failed to get or create user in optional auth")
				next.ServeHTTP(w, r)
				return
			}

			roles, err := roleProvider.GetRolesByUserId(userId)
			if err != nil {
				logger.WithError(err).Warn("Failed to load user roles in optional auth")
				next.ServeHTTP(w, r)
				return
			}

			authCtx := Context{
				UserId:   userId,
				Email:    headers.Email,
				Name:     headers.User,
				Provider: provider,
				Roles:    roles,
			}

			ctx := WithContext(r.Context(), authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates middleware that checks for a specific role
// Must be used after auth Middleware
func RequireRole(logger *logrus.Logger, role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := FromContext(r.Context())
			if !ok {
				logger.Warn("Auth context not found in RequireRole middleware")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !authCtx.HasRole(role) {
				logger.WithFields(logrus.Fields{
					"user_id":       authCtx.UserId.String(),
					"required_role": role,
					"user_roles":    authCtx.Roles,
				}).Warn("User does not have required role")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole creates middleware that checks for any of the specified roles
func RequireAnyRole(logger *logrus.Logger, roles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := FromContext(r.Context())
			if !ok {
				logger.Warn("Auth context not found in RequireAnyRole middleware")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !authCtx.HasAnyRole(roles) {
				logger.WithFields(logrus.Fields{
					"user_id":        authCtx.UserId.String(),
					"required_roles": roles,
					"user_roles":     authCtx.Roles,
				}).Warn("User does not have any required role")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SimpleUserProvider implements UserProvider using direct database access
type SimpleUserProvider struct {
	db *gorm.DB
}

// NewSimpleUserProvider creates a new SimpleUserProvider
func NewSimpleUserProvider(db *gorm.DB) *SimpleUserProvider {
	return &SimpleUserProvider{db: db}
}

// GetOrCreateUserByEmail gets an existing user or creates a new one
func (p *SimpleUserProvider) GetOrCreateUserByEmail(email, displayName, provider string) (uuid.UUID, error) {
	// Define a simple user entity for this operation
	type UserEntity struct {
		Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
		Email       string    `gorm:"type:varchar(255);not null;uniqueIndex"`
		DisplayName string    `gorm:"type:varchar(255);not null"`
		Provider    string    `gorm:"type:varchar(50);not null"`
		CreatedAt   time.Time `gorm:"not null"`
		UpdatedAt   time.Time `gorm:"not null"`
	}

	var user UserEntity

	// Try to find existing user
	err := p.db.Where("email = ?", email).First(&user).Error
	if err == nil {
		// User exists, update display name if changed
		if user.DisplayName != displayName {
			p.db.Model(&user).Update("display_name", displayName)
		}
		return user.Id, nil
	}

	if err != gorm.ErrRecordNotFound {
		return uuid.Nil, err
	}

	// User doesn't exist, create new one
	now := time.Now()
	user = UserEntity{
		Id:          uuid.New(),
		Email:       email,
		DisplayName: displayName,
		Provider:    provider,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := p.db.Create(&user).Error; err != nil {
		return uuid.Nil, err
	}

	// Assign default "user" role
	type RoleEntity struct {
		UserId uuid.UUID `gorm:"type:uuid;primaryKey"`
		Role   string    `gorm:"type:varchar(100);primaryKey"`
	}

	role := RoleEntity{
		UserId: user.Id,
		Role:   "user",
	}

	if err := p.db.Create(&role).Error; err != nil {
		// Log but don't fail if role assignment fails
		// The user is created, they just won't have a role yet
		return user.Id, nil
	}

	return user.Id, nil
}

// SimpleRoleProvider implements RoleProvider using direct database access
type SimpleRoleProvider struct {
	db *gorm.DB
}

// NewSimpleRoleProvider creates a new SimpleRoleProvider
func NewSimpleRoleProvider(db *gorm.DB) *SimpleRoleProvider {
	return &SimpleRoleProvider{db: db}
}

// GetRolesByUserId loads all roles for a given user
func (p *SimpleRoleProvider) GetRolesByUserId(userId uuid.UUID) ([]string, error) {
	type RoleEntity struct {
		UserId uuid.UUID `gorm:"type:uuid;primaryKey"`
		Role   string    `gorm:"type:varchar(100);primaryKey"`
	}

	var roles []RoleEntity
	if err := p.db.Where("user_id = ?", userId).Find(&roles).Error; err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Role
	}

	return roleNames, nil
}
