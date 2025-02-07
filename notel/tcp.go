package main

import (
	"log"
	"net"
	"sync"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/logs"
)

func serveTCP(wg *sync.WaitGroup, connConf confs.ConnectorConf, metrics *Metrics) {
	defer wg.Done()

	handler := func(conn net.Conn) {
		tcp, _ := minigo.NewTCP(conn)
		_ = tcp.Init()

		var innerWg sync.WaitGroup
		innerWg.Add(2)

		network := minigo.NewNetwork(tcp, false, &innerWg, "TCP")
		m := minigo.NewMinitel(network, false, &innerWg, connConf.Tag, metrics.ConnLostCount)
		m.NoCSI()
		go m.Serve()

		NotelApplication(m, &innerWg, &connConf, metrics)
		innerWg.Wait()

		logs.InfoLog("[%s] serve-TCP: disconnect\n", connConf.Tag)
		tcp.Disconnect()

		logs.InfoLog("[%s] serve-TCP: session closed\n", connConf.Tag)
	}

	err := listenAndServeTCP(connConf.Path, handler)
	log.Fatal(err)
}

func listenAndServeTCP(path string, handler func(net.Conn)) error {

	listener, err := net.Listen("tcp", path)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go handler(conn)
	}
}
