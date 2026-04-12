package retention

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds Prometheus collectors for the retention framework. They are
// labeled by service and category. Use NewMetrics(serviceName) and pass the
// returned struct to the Reaper.
type Metrics struct {
	service  string
	scanned  *prometheus.CounterVec
	deleted  *prometheus.CounterVec
	duration *prometheus.HistogramVec
	failures *prometheus.CounterVec
}

var (
	scannedVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "retention_scanned_total",
		Help: "Rows scanned by the retention reaper, partitioned by service and category.",
	}, []string{"service", "category"})

	deletedVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "retention_deleted_total",
		Help: "Rows deleted by the retention reaper, partitioned by service and category.",
	}, []string{"service", "category"})

	durationVec = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "retention_run_duration_seconds",
		Help:    "Wall-clock duration of one retention reaper run for a single (tenant, category).",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "category"})

	failuresVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "retention_run_failures_total",
		Help: "Reaper runs that ended in an error, partitioned by service and category.",
	}, []string{"service", "category"})
)

// NewMetrics returns a Metrics handle scoped to a service name.
func NewMetrics(service string) *Metrics {
	return &Metrics{
		service:  service,
		scanned:  scannedVec,
		deleted:  deletedVec,
		duration: durationVec,
		failures: failuresVec,
	}
}

// ObserveRun records counters and the duration histogram for one reaper run.
func (m *Metrics) ObserveRun(category Category, scanned, deleted int, durationSec float64, failed bool) {
	if m == nil {
		return
	}
	labels := prometheus.Labels{"service": m.service, "category": string(category)}
	m.scanned.With(labels).Add(float64(scanned))
	m.deleted.With(labels).Add(float64(deleted))
	m.duration.With(labels).Observe(durationSec)
	if failed {
		m.failures.With(labels).Inc()
	}
}

// Handler returns the HTTP handler that serves /metrics.
// Each service registers it via router.Handle("/metrics", retention.Handler()).
func Handler() http.Handler { return promhttp.Handler() }
