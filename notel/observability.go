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
	ConnectedUsers atomic.Int32

	ConnCount         *prometheus.CounterVec
	ConnAttemptCount  *prometheus.CounterVec
	ConnLostCount     *prometheus.CounterVec
	ConnDurationCount *prometheus.CounterVec
	ConnActive        *prometheus.GaugeVec

	MessagesCount prometheus.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		ConnCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "notel_connection_number",
			Help: "The total number connection to NOTEL",
		},
			[]string{"source"}),

		ConnAttemptCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "notel_connection_attempt_number",
			Help: "The total number of connection attempts to NOTEL",
		},
			[]string{"source"}),

		ConnLostCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "notel_connection_lost_number",
			Help: "The total number of lost connections on NOTEL",
		},
			[]string{"source"}),

		ConnActive: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "notel_connection_active",
			Help: "The number of currently active connections to NOTEL",
		},
			[]string{"source"}),

		ConnDurationCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "notel_connection_duration",
			Help: "The total connection duration to NOTEL",
		},
			[]string{"source"}),

		MessagesCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "notel_messages_number",
			Help: "The total number of postel messages to NOTEL",
		}),
	}
}

func serveMetrics(wg *sync.WaitGroup, metrics *Metrics, connectors []confs.ConnectorConf) {
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
