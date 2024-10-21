package main

import (
	"fmt"
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
		fmt.Errorf("notel: missing config file path\n")
		return
	}

	notelConfig, err := confs.LoadConfig(os.Args[1])
	if err != nil {
		fmt.Errorf("notel: unable to load conf: %s\n", err.Error())
		return
	}

	CommuneDb = databases.NewCommuneDatabase()
	CommuneDb.LoadCommuneDatabase(notelConfig.CommuneDbPath)

	MessageDb = databases.NewMessageDatabase()
	MessageDb.LoadMessages(notelConfig.MessagesDbPath)

	UsersDb = databases.NewUsersDatabase()
	UsersDb.LoadDatabase(notelConfig.UsersDbPath)

	for _, connConfig := range notelConfig.Connectors {
		if !connConfig.Active {
			continue
		}

		switch connConfig.Kind {
		case "modem":
			group.Add(1)
			go serveModem(&group, connConfig)
		case "websocket":
			group.Add(1)
			go serveWS(&group, connConfig)
		}
	}

	group.Add(1)
	go serverMetrics(&group, notelConfig.Connectors)

	group.Wait()

	MessageDb.Quit()
	UsersDb.Quit()
}

func NotelApplication(mntl *minigo.Minitel, sourceTag string, wg *sync.WaitGroup) {
	wg.Add(1)

	promConnNb.With(prometheus.Labels{"source": sourceTag}).Inc()
	active := NbConnectedUsers.Add(1)
	promConnActive.With(prometheus.Labels{"source": sourceTag}).Inc()

	logs.InfoLog("[%s] notel-handler: start handler, connected=%d\n", sourceTag, active)
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
		SommaireHandler(mntl, creds["login"])
	}

	promConnDur.With(prometheus.Labels{"source": sourceTag}).Add(time.Since(startConn).Seconds())

	active = NbConnectedUsers.Add(-1)
	promConnActive.With(prometheus.Labels{"source": sourceTag}).Dec()

	logs.InfoLog("[%s] notel-handler: quit handler, connected=%d\n", sourceTag, active)

	wg.Done()
}
