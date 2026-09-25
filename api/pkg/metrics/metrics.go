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

	PushEnqueued  = prometheus.NewCounter(prometheus.CounterOpts{Name: "anochat_push_enqueued_total", Help: "Web Push jobs enqueued"})
	PushDropped   = prometheus.NewCounter(prometheus.CounterOpts{Name: "anochat_push_dropped_total", Help: "Web Push jobs dropped because the queue was full"})
	PushSendTotal = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "anochat_push_send_total", Help: "Web Push delivery results"}, []string{"result"})
)

func Register() {
	prometheus.MustRegister(
		ActiveConnections,
		TotalUsers,
		NewRegistrations,
		QueueSize,
		MatchDuration,
		PushEnqueued,
		PushDropped,
		PushSendTotal,
	)
}
