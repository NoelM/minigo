package main

import (
	"log"
	"sync"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/logs"
)

func modemServe(wg *sync.WaitGroup, connConf confs.ConnectorConf, metrics *Metrics) {
	defer wg.Done()

	modem, err := minigo.NewModem(connConf.Path, 115200, connConf.Config, connConf.Tag, metrics.ConnAttemptCount)
	if err != nil {
		log.Fatal(err)
	}

	err = modem.Init()
	if err != nil {
		log.Fatal(err)
	}

	modem.RingHandler(func(modem *minigo.Modem) {
		var group sync.WaitGroup

		network := minigo.NewNetwork(modem, true, &group, connConf.Tag)
		minitel := minigo.NewMinitel(network, true, &group, connConf.Tag, metrics.ConnLostCount)
		go minitel.Serve()

		NotelApplication(minitel, &group, &connConf, metrics)
		group.Wait()

		logs.InfoLog("[%s] ring-handler: disconnect\n", connConf.Tag)
		modem.Disconnect()

		logs.InfoLog("[%s] ring-handler: minitel session closed\n", connConf.Tag)
	})

	modem.Serve(false)
}
