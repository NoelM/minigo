package main

import (
	"os"
	"sync"

	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/logs"
)

var CommuneDb *databases.CommuneDatabase
var MessageDb *databases.MessageDatabase
var UsersDb *databases.UsersDatabase
var BlogDbPath string
var AnnuaireDbPath string

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

	BlogDbPath = notelConf.BlogDbPath
	AnnuaireDbPath = notelConf.AnnuaireDbPath

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
