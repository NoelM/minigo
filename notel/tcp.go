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

func tcpServe(wg *sync.WaitGroup, connConf confs.ConnectorConf, metrics *Metrics) {
	defer wg.Done()

	handler := func(conn net.Conn) {
		tcp, _ := minigo.NewTCP(conn)
		_ = tcp.Init()

		number, err := getCallMetadata(tcp)
		if err != nil {
			logs.ErrorLog("[%s:<tel-number>] tcp-serve: unable to connect %s\n", connConf.Tag, err)
			return
		}

		fullTag := fmt.Sprintf("%s:%s", connConf.Tag, number)
		logs.InfoLog("[%s] tcp-serve: new call", fullTag)

		// Sleep while the connection is fully established, 7s is a approx.
		time.Sleep(7 * time.Second)

		var group sync.WaitGroup

		network := minigo.NewNetwork(tcp, false, &group, fullTag)
		minitel := minigo.NewMinitel(network, false, &group, fullTag, metrics.ConnLostCount)
		minitel.NoCSI()
		go minitel.Serve()

		NotelApplication(minitel, &group, &connConf, metrics)
		group.Wait()

		logs.InfoLog("[%s] tcp-serve: disconnect\n", fullTag)
		tcp.Disconnect()

		logs.InfoLog("[%s] tcp-serve: terminated\n", fullTag)
	}

	err := tcpListenAndServe(connConf.Path, handler)
	log.Fatal(err)
}

func tcpListenAndServe(path string, handler func(net.Conn)) error {

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

func getCallMetadata(tcp *minigo.TCP) (string, error) {
	/*
		CALLFROM 0173742367
		STARTURL
		PCE 1
	*/

	msg := make([]byte, 0)
	re := regexp.MustCompile(`CALLFROM\s(\d+)\nSTARTURL\s\nPCE\s\d`)

	for {
		part, err := tcp.Read()
		if err != nil {
			return "", err
		}

		msg = append(msg, part...)

		if re.Match(msg) {
			return re.FindStringSubmatch(string(msg))[1], nil
		}
	}
}
