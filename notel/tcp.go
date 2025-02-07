package main

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"sync"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/confs"
	"github.com/NoelM/minigo/notel/logs"
)

func serveTCP(wg *sync.WaitGroup, connConf confs.ConnectorConf, metrics *Metrics) {
	defer wg.Done()

	handler := func(conn net.Conn) {
		tcp, _ := minigo.NewTCP(conn)
		_ = tcp.Init()

		number, err := waitForConnect(tcp)
		if err != nil {
			logs.ErrorLog("[%s] serve-TCP: unable to connect %s\n", err)
			return
		}
		logs.InfoLog("[%s] serve-TCP: connect from number %s\n", connConf.Tag, number)

		fullTag := fmt.Sprintf("%s:%s", connConf.Tag, number)

		var innerWg sync.WaitGroup
		innerWg.Add(2)

		network := minigo.NewNetwork(tcp, false, &innerWg, "TCP")
		m := minigo.NewMinitel(network, false, &innerWg, fullTag, metrics.ConnLostCount)
		m.NoCSI()
		go m.Serve()

		NotelApplication(m, &innerWg, &connConf, metrics)
		innerWg.Wait()

		logs.InfoLog("[%s] serve-TCP: disconnect\n", fullTag)
		tcp.Disconnect()

		logs.InfoLog("[%s] serve-TCP: session closed\n", fullTag)
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

func waitForConnect(tcp *minigo.TCP) (string, error) {
	/*
		CALLFROM 0173742367
		STARTURL
		PCE 1
	*/

	msg := make([]byte, 0)
	re := regexp.MustCompile(`CALLFROM\s(\d+)\nSTARTURL\nPCE\s\d`)

	start := time.Now()
	for time.Since(start) < time.Minute {
		part, err := tcp.Read()
		if err != nil {
			return "", err
		}

		msg = append(msg, part...)

		if re.Match(msg) {
			return re.FindStringSubmatch(string(msg))[0], nil
		}
	}

	return "", fmt.Errorf("unable to connect V.23")
}
