package recipeimage

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collectors live on the default registry, which recipe-service already
// exports at /metrics.
var (
	uploadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "recipe_image_uploads_total",
		Help: "Recipe image uploads by result.",
	}, []string{"result"})

	processingSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "recipe_image_processing_seconds",
		Help:    "Time spent decoding, resizing and encoding one upload.",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
	})

	storageErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "recipe_image_storage_errors_total",
		Help: "Object storage failures by operation.",
	}, []string{"op"})
)

func recordUpload(result string) { uploadsTotal.WithLabelValues(result).Inc() }

// RecordStorageError increments recipe_image_storage_errors_total{op}.
func RecordStorageError(op string) { storageErrorsTotal.WithLabelValues(op).Inc() }
