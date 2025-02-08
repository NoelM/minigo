package main

import (
	"os"
	"sync"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/logs"

	"github.com/prometheus/client_golang/prometheus"
)

var CommuneDb *databases.CommuneDatabase
var MessageDb *databases.MessageDatabase
var UsersDb *databases.UsersDatabase

func main() {
	var group sync.WaitGroup

	if len(os.Args) == 1 {
		logs.ErrorLog("notel: missing config file path\n")
		return
	}

	notelConf, err := confs.LoadConfig(os.Args[1])
	if err != nil {
		logs.ErrorLog("notel: unable to load conf: %s\n", err.Error())
		return
	}

	CommuneDb = databases.NewCommuneDatabase()
	CommuneDb.LoadCommuneDatabase(notelConf.CommuneDbPath)

	MessageDb = databases.NewMessageDatabase()
	MessageDb.LoadMessages(notelConf.MessagesDbPath)

	UsersDb = databases.NewUsersDatabase()
	UsersDb.LoadDatabase(notelConf.UsersDbPath)

	group.Add(1)
	metrics := NewMetrics()
	go metricsServe(&group, metrics, notelConf.Connectors)

	for _, connConf := range notelConf.Connectors {
		if !connConf.Active {
			continue
		}

		switch connConf.Kind {
		case "modem":
			group.Add(1)
			go modemServe(&group, connConf, metrics)

		case "websocket":
			group.Add(1)
			go webSocketServe(&group, connConf, metrics)

		case "tcp":
			group.Add(1)
			go tcpServe(&group, connConf, metrics)
		}
	}
	group.Wait()

	MessageDb.Quit()
	UsersDb.Quit()
}

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
		SommaireHandler(minitel, creds["login"], metrics)
	}

	metrics.ConnDurationCount.With(prometheus.Labels{"source": connConf.Tag}).Add(time.Since(startConn).Seconds())

	activeUsers = metrics.ConnectedUsers.Add(-1)
	metrics.ConnActive.With(prometheus.Labels{"source": connConf.Tag}).Dec()

	logs.InfoLog("[%s] notel-handler: quit handler, connected=%d\n", connConf.Tag, activeUsers)

	group.Done()
}
