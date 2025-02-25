package main

import (
	"os"
	"sync"

	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/logs"
	"github.com/NoelM/minigo/notel/metrics"
)

var CommuneDb *databases.CommuneDatabase
var ChannelDb *databases.Channel
var UsersDb *databases.UsersDatabase
var BlogDbPath string
var AnnuaireDbPath string
var ChatManager *databases.ChatManager

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

	ChannelDb = databases.NewChannel()
	ChannelDb.LoadMessages(notelConf.ChannelsDb[0])

	ChatManager = databases.NewChatManager(notelConf)

	UsersDb = databases.NewUsersDatabase()
	UsersDb.LoadDatabase(notelConf.UsersDbPath)

	BlogDbPath = notelConf.BlogDbPath
	AnnuaireDbPath = notelConf.AnnuaireDbPath

	group.Add(1)
	mtr := metrics.NewMetrics()
	go metrics.Serve(&group, mtr, notelConf.Connectors)

	for _, connConf := range notelConf.Connectors {
		if !connConf.Active {
			continue
		}

		switch connConf.Kind {
		case "modem":
			group.Add(1)
			go modemServe(&group, connConf, mtr)

		case "websocket":
			group.Add(1)
			go webSocketServe(&group, connConf, mtr)

		case "tcp":
			group.Add(1)
			go tcpServe(&group, connConf, mtr)
		}
	}
	group.Wait()

	ChannelDb.Quit()
	UsersDb.Quit()
}
