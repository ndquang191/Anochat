package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ActiveConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "anochat_active_ws_connections",
		Help: "Number of active WebSocket connections",
	})

	TotalUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "anochat_users_total",
		Help: "Total number of registered users",
	})

	NewRegistrations = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "anochat_new_registrations_total",
		Help: "Total new user registrations since server start",
	})

	QueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "anochat_queue_size",
		Help: "Number of users currently waiting in the matching queue",
	})

	MatchDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "anochat_match_duration_seconds",
		Help:    "Time in seconds from joining queue to being matched",
		Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120, 300},
	})
)

func Register() {
	prometheus.MustRegister(
		ActiveConnections,
		TotalUsers,
		NewRegistrations,
		QueueSize,
		MatchDuration,
	)
}
