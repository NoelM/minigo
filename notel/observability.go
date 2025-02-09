package main

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/NoelM/minigo/notel/confs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	mutex sync.RWMutex

	ConnectedUsers atomic.Int32
	loggedUsers    map[string]bool

	ConnCount         *prometheus.CounterVec
	ConnAttemptCount  *prometheus.CounterVec
	ConnLostCount     *prometheus.CounterVec
	ConnDurationCount *prometheus.CounterVec
	ConnActive        *prometheus.GaugeVec

	MessagesCount prometheus.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		loggedUsers: make(map[string]bool),

		ConnCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "notel_connection_total",
			Help: "Count of successful connections.",
		},
			[]string{"source"}),

		ConnAttemptCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "notel_connection_attempt_total",
			Help: "Count of connection attemps.",
		},
			[]string{"source"}),

		ConnLostCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "notel_connection_lost_total",
			Help: "Count of lost connections.",
		},
			[]string{"source"}),

		ConnActive: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "notel_connection_active",
			Help: "Gauge of active connections.",
		},
			[]string{"source"}),

		ConnDurationCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "notel_connection_duration_seconds_total",
			Help: "Count of connection durations in seconds.",
		},
			[]string{"source"}),

		MessagesCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "notel_messages_total",
			Help: "Count of posted messages.",
		}),
	}
}

func (m *Metrics) Logged(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.loggedUsers[name] = true
}

func (m *Metrics) Disconnect(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.loggedUsers, name)
}

func (m *Metrics) ListLogged() []string {
	m.mutex.RLock()
	defer m.mutex.Unlock()

	keys := make([]string, 0, len(m.loggedUsers))
	for k := range m.loggedUsers {
		keys = append(keys, k)
	}

	return keys
}

func (m *Metrics) CountLogged() int {
	m.mutex.RLock()
	defer m.mutex.Unlock()

	return len(m.loggedUsers)
}

func metricsServe(wg *sync.WaitGroup, metrics *Metrics, connectors []confs.ConnectorConf) {
	defer wg.Done()

	for _, cv := range []*prometheus.CounterVec{metrics.ConnCount, metrics.ConnLostCount, metrics.ConnDurationCount, metrics.ConnAttemptCount} {
		for _, connConf := range connectors {
			if !connConf.Active {
				continue
			}
			cv.With(prometheus.Labels{"source": connConf.Tag}).Inc()
		}
	}

	for _, connConf := range connectors {
		if !connConf.Active {
			continue
		}
		metrics.ConnActive.With(prometheus.Labels{"source": connConf.Tag}).Set(0)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
