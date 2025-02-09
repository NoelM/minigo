package main

import (
	"sync"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/prometheus/client_golang/prometheus"
)

func NotelApplication(minitel *minigo.Minitel, group *sync.WaitGroup, connConf *confs.ConnectorConf, metrics *Metrics) {
	group.Add(1)

	metrics.ConnCount.With(prometheus.Labels{"source": connConf.Tag}).Inc()

	activeUsers := metrics.ConnectedUsers.Add(1)
	metrics.ConnActive.With(prometheus.Labels{"source": connConf.Tag}).Inc()

	logs.InfoLog("[%s] notel-handler: start handler, connected=%d\n", connConf.Tag, activeUsers)
	startConn := time.Now()

SIGNIN:
	creds, op := NewPageSignIn(minitel).Run()

	if op == minigo.GuideOp {
		creds, op = NewSignUpPage(minitel).Run()

		if op == minigo.SommaireOp {
			goto SIGNIN
		}
	}

	if op == minigo.EnvoiOp {
		metrics.Logged(creds["login"])
		SommaireHandler(minitel, creds["login"], metrics)
	}

	metrics.ConnDurationCount.With(prometheus.Labels{"source": connConf.Tag}).Add(time.Since(startConn).Seconds())

	activeUsers = metrics.ConnectedUsers.Add(-1)
	metrics.ConnActive.With(prometheus.Labels{"source": connConf.Tag}).Dec()

	if login, ok := creds["login"]; ok {
		metrics.Disconnect(login)
	}

	logs.InfoLog("[%s] notel-handler: quit handler, connected=%d\n", connConf.Tag, activeUsers)

	group.Done()
}
