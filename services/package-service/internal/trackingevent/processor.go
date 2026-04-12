package trackingevent

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) GetByPackageID(packageID uuid.UUID) model.Provider[[]Model] {
	return func() ([]Model, error) {
		entities, err := GetByPackageID(packageID)(p.db.WithContext(p.ctx))()
		if err != nil {
			return nil, err
		}
		models := make([]Model, 0, len(entities))
		for _, e := range entities {
			m, merr := Make(e)
			if merr != nil {
				return nil, merr
			}
			models = append(models, m)
		}
		return models, nil
	}
}

func (p *Processor) CreateEvent(packageID uuid.UUID, timestamp time.Time, status, description string, location, rawStatus *string) error {
	return Create(p.db.WithContext(p.ctx), packageID, timestamp, status, description, location, rawStatus)
}
