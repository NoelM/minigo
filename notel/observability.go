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

var NbConnectedUsers atomic.Int32

var (
	promConnNb = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notel_connection_number",
		Help: "The total number connection to NOTEL",
	},
		[]string{"source"})

	promConnAttemptNb = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notel_connection_attempt_number",
		Help: "The total number of connection attempts to NOTEL",
	},
		[]string{"source"})

	promConnLostNb = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notel_connection_lost_number",
		Help: "The total number of lost connections on NOTEL",
	},
		[]string{"source"})

	promConnActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "notel_connection_active",
		Help: "The number of currently active connections to NOTEL",
	},
		[]string{"source"})

	promConnDur = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notel_connection_duration",
		Help: "The total connection duration to NOTEL",
	},
		[]string{"source"})

	promMsgNb = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notel_messages_number",
		Help: "The total number of postel messages to NOTEL",
	})
)

func serverMetrics(wg *sync.WaitGroup, connectors []confs.ConnectorConf) {
	defer wg.Done()

	for _, cv := range []*prometheus.CounterVec{promConnNb, promConnLostNb, promConnDur, promConnAttemptNb} {
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
		promConnActive.With(prometheus.Labels{"source": connConf.Tag}).Set(0)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
