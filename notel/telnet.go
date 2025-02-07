package main

import (
	"log"
	"net"
	"sync"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/logs"
)

func serveTelnet(wg *sync.WaitGroup, connConf confs.ConnectorConf, metrics *Metrics) {
	defer wg.Done()

	listener, err := net.Listen("tcp", connConf.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			break
		}

		tel, _ := minigo.NewTelnet(conn)
		_ = tel.Init()

		var innerWg sync.WaitGroup
		innerWg.Add(2)

		network := minigo.NewNetwork(tel, false, &innerWg, "telnet")
		m := minigo.NewMinitel(network, false, &innerWg, connConf.Tag, metrics.ConnLostCount)
		m.NoCSI()
		go m.Serve()

		NotelApplication(m, &innerWg, &connConf, metrics)
		innerWg.Wait()

		logs.InfoLog("[%s] serve-telnet: disconnect\n", connConf.Tag)
		tel.Disconnect()

		logs.InfoLog("[%s] serve-telnet: session closed\n", connConf.Tag)
	}

	log.Fatal(err)
}
