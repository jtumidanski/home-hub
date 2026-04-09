package geocoding

import (
	"context"

	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/sirupsen/logrus"
)

type Processor struct {
	l      logrus.FieldLogger
	ctx    context.Context
	client *openmeteo.Client
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, client *openmeteo.Client) *Processor {
	return &Processor{l: l, ctx: ctx, client: client}
}

func (p *Processor) Search(query string) ([]openmeteo.GeocodingResult, error) {
	return p.client.SearchPlaces(query)
}
