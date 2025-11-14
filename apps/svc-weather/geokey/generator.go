package geokey

import (
	"github.com/mmcloughlin/geohash"
)

// Generator provides geohash-based spatial key generation
type Generator struct {
	precision int
}

// NewGenerator creates a new Generator with the specified precision
// Precision of 5 provides ~5km accuracy
func NewGenerator(precision int) Generator {
	return Generator{
		precision: precision,
	}
}

// Generate creates a geohash key from latitude and longitude
func (g Generator) Generate(lat, lon float64) string {
	return geohash.EncodeWithPrecision(lat, lon, uint(g.precision))
}

// Precision returns the configured precision level
func (g Generator) Precision() int {
	return g.precision
}
