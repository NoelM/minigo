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

	notelConfig, err := confs.LoadConfig(os.Args[1])
	if err != nil {
		logs.ErrorLog("notel: unable to load conf: %s\n", err.Error())
		return
	}

	CommuneDb = databases.NewCommuneDatabase()
	CommuneDb.LoadCommuneDatabase(notelConfig.CommuneDbPath)

	MessageDb = databases.NewMessageDatabase()
	MessageDb.LoadMessages(notelConfig.MessagesDbPath)

	UsersDb = databases.NewUsersDatabase()
	UsersDb.LoadDatabase(notelConfig.UsersDbPath)

	group.Add(1)
	metrics := NewMetrics()
	go serveMetrics(&group, metrics, notelConfig.Connectors)

	for _, connConfig := range notelConfig.Connectors {
		if !connConfig.Active {
			continue
		}

		switch connConfig.Kind {
		case "modem":
			group.Add(1)
			go serveModem(&group, connConfig, metrics)

		case "websocket":
			group.Add(1)
			go serveWebSocket(&group, connConfig, metrics)

		case "telnet":
			group.Add(1)
			go serveTelnet(&group, connConfig, metrics)
		}
	}
	group.Wait()

	MessageDb.Quit()
	UsersDb.Quit()
}

func NotelApplication(mntl *minigo.Minitel, wg *sync.WaitGroup, connConf *confs.ConnectorConf, metrics *Metrics) {
	wg.Add(1)

	metrics.ConnCount.With(prometheus.Labels{"source": connConf.Tag}).Inc()

	activeUsers := metrics.ConnectedUsers.Add(1)
	metrics.ConnActive.With(prometheus.Labels{"source": connConf.Tag}).Inc()

	logs.InfoLog("[%s] notel-handler: start handler, connected=%d\n", connConf.Tag, activeUsers)
	startConn := time.Now()

SIGNIN:
	creds, op := NewPageSignIn(mntl).Run()

	if op == minigo.GuideOp {
		creds, op = NewSignUpPage(mntl).Run()

		if op == minigo.SommaireOp {
			goto SIGNIN
		}
	}

	if op == minigo.EnvoiOp {
		SommaireHandler(mntl, creds["login"], metrics)
	}

	metrics.ConnDurationCount.With(prometheus.Labels{"source": connConf.Tag}).Add(time.Since(startConn).Seconds())

	activeUsers = metrics.ConnectedUsers.Add(-1)
	metrics.ConnActive.With(prometheus.Labels{"source": connConf.Tag}).Dec()

	logs.InfoLog("[%s] notel-handler: quit handler, connected=%d\n", connConf.Tag, activeUsers)

	wg.Done()
}
