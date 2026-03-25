package refreshtoken

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, userID uuid.UUID, tokenHash string, expiresAt time.Time) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:        uuid.New(),
		UserId:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Revoked:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func revokeByHash(db *gorm.DB, hash string) error {
	return db.Model(&Entity{}).
		Where("token_hash = ?", hash).
		Update("revoked", true).Error
}

func revokeAllForUser(db *gorm.DB, userID uuid.UUID) error {
	return db.Model(&Entity{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}
